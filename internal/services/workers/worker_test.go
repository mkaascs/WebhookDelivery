package workers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"webhook-delivery/internal/config"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	fixedTime  = time.Date(2026, time.July, 1, 12, 0, 0, 0, time.UTC)
	deliveryID = uuid.MustParse("01900000-0000-7000-8000-000000000001")
)

func newTestWorker(t *testing.T, cfg config.Workers) (*Worker, *MockDeliveryRepo) {
	ctrl := gomock.NewController(t)
	repo := NewMockDeliveryRepo(ctrl)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewWorkers(repo, cfg, log), repo
}

func testWorkersConfig() config.Workers {
	return config.Workers{
		MaxGoroutines: 1,
		MaxAttempts:   5,
		BaseBackoff:   time.Second,
		MaxBackoff:    time.Hour,
		Timeout:       5 * time.Second,
		QueueSize:     2,
	}
}

func newPendingDelivery(url string, attempts int) domain.PendingDelivery {
	return domain.PendingDelivery{
		Delivery: domain.Delivery{
			ID:          deliveryID,
			Status:      domain.StatusProcessing,
			Attempts:    attempts,
			MaxAttempts: 5,
			NextRetryAt: fixedTime,
		},
		URL:     url,
		Secret:  "s",
		Payload: json.RawMessage(`{}`),
	}
}

func Test_isSuccessCode(t *testing.T) {
	tests := []struct {
		code int
		want bool
	}{
		{0, false},
		{199, false},
		{200, true},
		{201, true},
		{299, true},
		{300, true},
		{399, true},
		{400, false},
		{404, false},
		{500, false},
	}

	for _, tt := range tests {
		require.Equal(t, tt.want, isSuccessCode(tt.code), "code %d", tt.code)
	}
}

func Test_sendPostRequest(t *testing.T) {
	payload := json.RawMessage(`{"amount":100}`)
	secret := []byte("whsec_top")

	t.Run("success sends signed request and returns status", func(t *testing.T) {
		var gotSignature, gotContentType string
		var gotBody []byte
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotSignature = r.Header.Get("X-Webhook-Signature")
			gotContentType = r.Header.Get("Content-Type")
			gotBody, _ = io.ReadAll(r.Body)
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		worker, _ := newTestWorker(t, testWorkersConfig())
		code, desc, err := worker.sendPostRequest(context.Background(), srv.URL, payload, secret)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, code)
		require.Equal(t, "HTTP 200", desc)

		mac := hmac.New(sha256.New, secret)
		mac.Write(payload)
		require.Equal(t, hex.EncodeToString(mac.Sum(nil)), gotSignature)
		require.Equal(t, "application/json", gotContentType)
		require.Equal(t, []byte(payload), gotBody)
	})

	t.Run("non-2xx status is returned without error", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		worker, _ := newTestWorker(t, testWorkersConfig())
		code, desc, err := worker.sendPostRequest(context.Background(), srv.URL, payload, secret)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, code)
		require.Equal(t, "HTTP 500", desc)
	})

	t.Run("transport error returns error and zero code", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		url := srv.URL
		srv.Close()

		worker, _ := newTestWorker(t, testWorkersConfig())
		code, desc, err := worker.sendPostRequest(context.Background(), url, payload, secret)
		require.Error(t, err)
		require.Zero(t, code)
		require.Equal(t, err.Error(), desc)
	})
}

func Test_process(t *testing.T) {
	tests := []struct {
		name            string
		serverCode      int // 0 means the receiver is unreachable
		attempts        int // attempts before this one
		wantStatus      domain.DeliveryStatus
		wantAttempts    int
		wantRescheduled bool // pending gets a fresh next_retry_at, delivered and failed keep the original
		wantLastCode    *int
		wantLastError   bool
	}{
		{
			name:         "2xx -> delivered",
			serverCode:   http.StatusOK,
			wantStatus:   domain.StatusDelivered,
			wantAttempts: 1,
		},
		{
			name:         "3xx -> delivered",
			serverCode:   http.StatusNotModified,
			wantStatus:   domain.StatusDelivered,
			wantAttempts: 1,
		},
		{
			name:            "5xx with attempts left -> pending",
			serverCode:      http.StatusInternalServerError,
			wantStatus:      domain.StatusPending,
			wantAttempts:    1,
			wantRescheduled: true,
			wantLastCode:    new(http.StatusInternalServerError),
			wantLastError:   true,
		},
		{
			name:            "4xx with attempts left -> pending",
			serverCode:      http.StatusNotFound,
			attempts:        2,
			wantStatus:      domain.StatusPending,
			wantAttempts:    3,
			wantRescheduled: true,
			wantLastCode:    new(http.StatusNotFound),
			wantLastError:   true,
		},
		{
			name:          "last attempt failed -> failed",
			serverCode:    http.StatusInternalServerError,
			attempts:      4,
			wantStatus:    domain.StatusFailed,
			wantAttempts:  5,
			wantLastCode:  new(http.StatusInternalServerError),
			wantLastError: true,
		},
		{
			name:            "transport error -> pending with zero code",
			wantStatus:      domain.StatusPending,
			wantAttempts:    1,
			wantRescheduled: true,
			wantLastCode:    new(0),
			wantLastError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.serverCode)
			}))
			url := srv.URL
			if tt.serverCode == 0 {
				srv.Close()
			} else {
				defer srv.Close()
			}

			worker, repo := newTestWorker(t, testWorkersConfig())

			var got dto.UpdateDeliveryStatusCommand
			repo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any()).
				DoAndReturn(func(_ context.Context, cmd dto.UpdateDeliveryStatusCommand) error {
					got = cmd
					return nil
				})

			before := time.Now()
			worker.process(context.Background(), newPendingDelivery(url, tt.attempts))

			require.Equal(t, deliveryID, got.ID)
			require.Equal(t, tt.wantStatus, got.Status)
			require.Equal(t, tt.wantAttempts, got.Attempts)
			require.Equal(t, tt.wantLastCode, got.LastResponseCode)
			require.Equal(t, tt.wantLastError, got.LastError != nil)

			if tt.wantRescheduled {
				require.True(t, got.NextRetryAt.After(before))
			} else {
				require.Equal(t, fixedTime, got.NextRetryAt)
			}
		})
	}

	t.Run("update error is only logged", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()

		worker, repo := newTestWorker(t, testWorkersConfig())
		repo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any()).Return(errors.New("db down"))

		require.NotPanics(t, func() {
			worker.process(context.Background(), newPendingDelivery(srv.URL, 0))
		})
	})
}

func Test_backoff(t *testing.T) {
	cfg := testWorkersConfig()
	worker, _ := newTestWorker(t, cfg)

	tests := []struct {
		name     string
		attempts int
		base     time.Duration
	}{
		{name: "first attempt uses base backoff", attempts: 1, base: cfg.BaseBackoff},
		{name: "doubles every attempt", attempts: 4, base: 8 * cfg.BaseBackoff},
		{name: "capped by max backoff", attempts: 30, base: cfg.MaxBackoff},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for range 100 {
				got := worker.backoff(tt.attempts)
				require.GreaterOrEqual(t, got, time.Duration(float64(tt.base)*0.8))
				require.LessOrEqual(t, got, time.Duration(float64(tt.base)*1.2))
			}
		})
	}
}

func Test_EnqueueAndFreeSlots(t *testing.T) {
	worker, _ := newTestWorker(t, testWorkersConfig())
	require.Equal(t, 2, worker.FreeSlots())

	worker.Enqueue(newPendingDelivery("http://example.com", 0))
	require.Equal(t, 1, worker.FreeSlots())

	worker.Enqueue(newPendingDelivery("http://example.com", 0))
	require.Zero(t, worker.FreeSlots())
}

func Test_Run(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	worker, repo := newTestWorker(t, testWorkersConfig())

	processed := make(chan dto.UpdateDeliveryStatusCommand, 1)
	repo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, cmd dto.UpdateDeliveryStatusCommand) error {
			processed <- cmd
			return nil
		})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		worker.Run(ctx)
		close(done)
	}()

	worker.Enqueue(newPendingDelivery(srv.URL, 0))

	select {
	case cmd := <-processed:
		require.Equal(t, domain.StatusDelivered, cmd.Status)
	case <-time.After(5 * time.Second):
		t.Fatal("queued delivery was not processed")
	}

	cancel()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not stop after context cancel")
	}
}
