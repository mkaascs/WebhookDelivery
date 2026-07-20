package workers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"webhook-delivery/internal/config"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
)

func newTestService(t *testing.T, cfg config.WorkersConfig) (*Service, *MockDeliveryRepo) {
	ctrl := gomock.NewController(t)
	repo := NewMockDeliveryRepo(ctrl)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewService(repo, log, cfg), repo
}

func testConfig() config.WorkersConfig {
	return config.WorkersConfig{
		MaxGoroutines:  1,
		BatchSize:      10,
		MaxAttempts:    5,
		BaseBackoff:    time.Second,
		MaxBackoff:     time.Hour,
		TickerDuration: time.Minute,
		Timeout:        5 * time.Second,
	}
}

func Test_isSuccessCode(t *testing.T) {
	tests := []struct {
		code int
		want bool
	}{
		{200, true},
		{201, true},
		{299, true},
		{199, false},
		{300, false},
		{404, false},
		{500, false},
	}

	for _, tt := range tests {
		require.Equal(t, tt.want, isSuccessCode(tt.code))
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

		svc, _ := newTestService(t, testConfig())
		code, desc, err := svc.sendPostRequest(context.Background(), srv.URL, payload, secret)
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
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		svc, _ := newTestService(t, testConfig())
		code, desc, err := svc.sendPostRequest(context.Background(), srv.URL, payload, secret)
		require.NoError(t, err)
		require.Equal(t, http.StatusInternalServerError, code)
		require.Equal(t, "HTTP 500", desc)
	})

	t.Run("transport error returns error and zero code", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		url := srv.URL
		srv.Close()

		svc, _ := newTestService(t, testConfig())
		code, desc, err := svc.sendPostRequest(context.Background(), url, payload, secret)
		require.Error(t, err)
		require.Zero(t, code)
		require.Equal(t, err.Error(), desc)
	})
}

func Test_processBatch(t *testing.T) {
	fixedTime := time.Date(2026, time.July, 1, 12, 0, 0, 0, time.UTC)

	newDelivery := func(url string) dto.ClaimPendingResult {
		return dto.ClaimPendingResult{
			ID:          "del-1",
			URL:         url,
			Secret:      []byte("s"),
			Payload:     json.RawMessage(`{}`),
			Attempts:    0,
			MaxAttempts: 5,
			NextRetryAt: fixedTime,
		}
	}

	t.Run("claim error skips processing", func(t *testing.T) {
		svc, repo := newTestService(t, testConfig())
		repo.EXPECT().ClaimPending(gomock.Any(), gomock.Any()).Return(nil, context.DeadlineExceeded)

		svc.processBatch(context.Background())
	})

	t.Run("empty batch does nothing", func(t *testing.T) {
		svc, repo := newTestService(t, testConfig())
		repo.EXPECT().ClaimPending(gomock.Any(), gomock.Any()).Return(nil, nil)

		svc.processBatch(context.Background())
	})

	t.Run("update status by response code and attempts", func(t *testing.T) {
		cases := []struct {
			name            string
			serverCode      int
			attempts        int
			maxAttempts     int
			wantStatus      domain.DeliveryStatus
			wantAttempts    int
			wantRescheduled bool // pending gets a fresh next_retry_at; delivered/failed keep the original
		}{
			{"2xx -> delivered", http.StatusOK, 0, 5, domain.StatusDelivered, 1, false},
			{"non-2xx with attempts left -> pending", http.StatusInternalServerError, 0, 5, domain.StatusPending, 1, true},
			{"non-2xx with attempts exhausted -> failed", http.StatusInternalServerError, 4, 5, domain.StatusFailed, 5, false},
		}

		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					w.WriteHeader(c.serverCode)
				}))
				defer srv.Close()

				svc, repo := newTestService(t, testConfig())
				delivery := newDelivery(srv.URL)
				delivery.Attempts = c.attempts
				delivery.MaxAttempts = c.maxAttempts
				repo.EXPECT().ClaimPending(gomock.Any(), gomock.Any()).
					Return([]dto.ClaimPendingResult{delivery}, nil)

				var got dto.UpdateDeliveryStatusCommand
				repo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, cmd dto.UpdateDeliveryStatusCommand) error {
						got = cmd
						return nil
					})

				before := time.Now()
				svc.processBatch(context.Background())

				require.Equal(t, delivery.ID, got.ID)
				require.Equal(t, c.wantStatus, got.Status)
				require.Equal(t, c.wantAttempts, got.Attempts)

				require.NotNil(t, got.LastResponseCode)
				require.Equal(t, c.serverCode, *got.LastResponseCode)

				if c.wantStatus == domain.StatusDelivered {
					require.Nil(t, got.LastError)
					require.Equal(t, fixedTime, got.NextRetryAt)
					return
				}

				require.NotNil(t, got.LastError)
				require.Equal(t, fmt.Sprintf("HTTP %d", c.serverCode), *got.LastError)
				if c.wantRescheduled {
					require.True(t, got.NextRetryAt.After(before))
				} else {
					require.Equal(t, fixedTime, got.NextRetryAt)
				}
			})
		}
	})

	t.Run("transport error is retried as pending", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
		url := srv.URL
		srv.Close()

		svc, repo := newTestService(t, testConfig())
		repo.EXPECT().ClaimPending(gomock.Any(), gomock.Any()).
			Return([]dto.ClaimPendingResult{newDelivery(url)}, nil)

		var got dto.UpdateDeliveryStatusCommand
		repo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, cmd dto.UpdateDeliveryStatusCommand) error {
				got = cmd
				return nil
			})

		before := time.Now()
		svc.processBatch(context.Background())

		require.Equal(t, domain.StatusPending, got.Status)
		require.Equal(t, 1, got.Attempts)
		require.True(t, got.NextRetryAt.After(before))
		require.Nil(t, got.LastResponseCode)
		require.NotNil(t, got.LastError)
	})
}
