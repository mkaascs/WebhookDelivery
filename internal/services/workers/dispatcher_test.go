package workers

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"webhook-delivery/internal/config"
	"webhook-delivery/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func newTestDispatcher(t *testing.T, cfg config.Dispatcher) (*Dispatcher, *MockQueue, *MockDispatcherRepo) {
	ctrl := gomock.NewController(t)
	queue := NewMockQueue(ctrl)
	repo := NewMockDispatcherRepo(ctrl)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewDispatcher(queue, repo, cfg, log), queue, repo
}

func Test_dispatch(t *testing.T) {
	first := &domain.PendingDelivery{Delivery: domain.Delivery{ID: uuid.MustParse("01900000-0000-7000-8000-000000000001")}}
	second := &domain.PendingDelivery{Delivery: domain.Delivery{ID: uuid.MustParse("01900000-0000-7000-8000-000000000002")}}

	tests := []struct {
		name      string
		batchSize int
		freeSlots int
		wantLimit int
		claimed   []*domain.PendingDelivery
		claimErr  error
	}{
		{
			name:      "limit is capped by batch size",
			batchSize: 10,
			freeSlots: 25,
			wantLimit: 10,
			claimed:   []*domain.PendingDelivery{first, second},
		},
		{
			name:      "limit is capped by free queue slots",
			batchSize: 10,
			freeSlots: 3,
			wantLimit: 3,
			claimed:   []*domain.PendingDelivery{first},
		},
		{
			name:      "nothing to claim",
			batchSize: 10,
			freeSlots: 10,
			wantLimit: 10,
		},
		{
			name:      "claim error enqueues nothing",
			batchSize: 10,
			freeSlots: 10,
			wantLimit: 10,
			claimErr:  errors.New("db down"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dispatcher, queue, repo := newTestDispatcher(t, config.Dispatcher{BatchSize: tt.batchSize})

			queue.EXPECT().FreeSlots().Return(tt.freeSlots)
			repo.EXPECT().ClaimPending(gomock.Any(), tt.wantLimit).Return(tt.claimed, tt.claimErr)

			var calls []any
			for _, delivery := range tt.claimed {
				calls = append(calls, queue.EXPECT().Enqueue(delivery))
			}
			gomock.InOrder(calls...)

			dispatcher.dispatch(context.Background())
		})
	}
}

func Test_Recover(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		dispatcher, _, repo := newTestDispatcher(t, config.Dispatcher{})
		repo.EXPECT().RecoverProcessing(gomock.Any()).Return(nil)

		require.NoError(t, dispatcher.Recover(context.Background()))
	})

	t.Run("error is wrapped", func(t *testing.T) {
		dispatcher, _, repo := newTestDispatcher(t, config.Dispatcher{})
		repoErr := errors.New("db down")
		repo.EXPECT().RecoverProcessing(gomock.Any()).Return(repoErr)

		require.ErrorIs(t, dispatcher.Recover(context.Background()), repoErr)
	})
}

func Test_Dispatcher_Run(t *testing.T) {
	dispatcher, queue, repo := newTestDispatcher(t, config.Dispatcher{Interval: 10 * time.Millisecond, BatchSize: 10})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	queue.EXPECT().FreeSlots().Return(10).MinTimes(1)
	repo.EXPECT().ClaimPending(gomock.Any(), 10).
		DoAndReturn(func(context.Context, int) ([]*domain.PendingDelivery, error) {
			cancel()
			return nil, nil
		}).MinTimes(1)

	done := make(chan struct{})
	go func() {
		dispatcher.Run(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not stop after context cancel")
	}
}
