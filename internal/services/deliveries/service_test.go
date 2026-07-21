package deliveries

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
)

func newTestService(t *testing.T) (*Service, *MockDeliveryRepo, *MockRetryNotifier) {
	ctrl := gomock.NewController(t)
	repo := NewMockDeliveryRepo(ctrl)
	notifier := NewMockRetryNotifier(ctrl)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewService(repo, notifier, log), repo, notifier
}

func Test_Service_GetByID(t *testing.T) {
	const id = "del-1"
	delivery := &domain.Delivery{ID: id, Status: domain.StatusDelivered}

	t.Run("success", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		repo.EXPECT().GetByID(gomock.Any(), id).Return(delivery, nil)

		got, err := svc.GetByID(context.Background(), id)
		require.NoError(t, err)
		require.Equal(t, &dto.GetDeliveryResult{Delivery: *delivery}, got)
	})

	t.Run("error is wrapped", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		repo.EXPECT().GetByID(gomock.Any(), id).Return(nil, domain.ErrDeliveryNotFound)

		got, err := svc.GetByID(context.Background(), id)
		require.Nil(t, got)
		require.ErrorIs(t, err, domain.ErrDeliveryNotFound)
	})
}

func Test_Service_GetFromEvent(t *testing.T) {
	const eventID = "ev-1"
	deliveries := []domain.Delivery{
		{ID: "del-1", EventID: eventID, Status: domain.StatusPending},
		{ID: "del-2", EventID: eventID, Status: domain.StatusFailed},
	}
	want := &dto.GetDeliveriesFromEventResult{
		Total:      12,
		Deliveries: deliveries,
	}

	t.Run("success", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		repo.EXPECT().GetFromEvent(gomock.Any(), dto.GetDeliveriesFromEventCommand{
			EventID: eventID,
		}).Return(deliveries, 12, nil)

		got, err := svc.GetFromEvent(context.Background(), dto.GetDeliveriesFromEventCommand{
			EventID: eventID,
		})

		require.NoError(t, err)
		require.Equal(t, want, got)
	})

	t.Run("error is wrapped", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		repo.EXPECT().GetFromEvent(gomock.Any(), dto.GetDeliveriesFromEventCommand{
			EventID: eventID,
		}).Return(nil, 0, domain.ErrEventNotFound)

		got, err := svc.GetFromEvent(context.Background(), dto.GetDeliveriesFromEventCommand{
			EventID: eventID,
		})

		require.Nil(t, got)
		require.ErrorIs(t, err, domain.ErrEventNotFound)
	})
}

func Test_Service_Retry(t *testing.T) {
	const id = "del-1"

	t.Run("success resets the delivery to pending", func(t *testing.T) {
		svc, repo, notifier := newTestService(t)

		var got dto.UpdateDeliveryStatusCommand
		repo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, cmd dto.UpdateDeliveryStatusCommand) error {
				got = cmd
				return nil
			})

		notifier.EXPECT().Notify()

		before := time.Now()
		require.NoError(t, svc.Retry(context.Background(), id))

		require.Equal(t, id, got.ID)
		require.Equal(t, domain.StatusPending, got.Status)
		require.Equal(t, 0, got.Attempts)
		require.Nil(t, got.LastError)
		require.Nil(t, got.LastResponseCode)
		require.False(t, got.NextRetryAt.Before(before))
	})

	t.Run("error is wrapped", func(t *testing.T) {
		svc, repo, _ := newTestService(t)
		repo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any()).Return(domain.ErrDeliveryNotFound)

		require.ErrorIs(t, svc.Retry(context.Background(), id), domain.ErrDeliveryNotFound)
	})
}
