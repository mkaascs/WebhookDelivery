package deliveries

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
)

func newTestService(t *testing.T) (*Service, *MockDeliveryRepo) {
	ctrl := gomock.NewController(t)
	repo := NewMockDeliveryRepo(ctrl)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewService(repo, log), repo
}

var (
	deliveryID  = uuid.MustParse("01900000-0000-7000-8000-000000000001")
	deliveryID2 = uuid.MustParse("01900000-0000-7000-8000-000000000002")
	eventID     = uuid.MustParse("01900000-0000-7000-8000-0000000000e1")
)

func Test_Service_GetByID(t *testing.T) {
	id := deliveryID
	delivery := &domain.Delivery{ID: id, Status: domain.StatusDelivered}

	t.Run("success", func(t *testing.T) {
		svc, repo := newTestService(t)
		repo.EXPECT().GetByID(gomock.Any(), id).Return(delivery, nil)

		got, err := svc.GetByID(context.Background(), id)
		require.NoError(t, err)
		require.Equal(t, delivery, got)
	})

	t.Run("error is wrapped", func(t *testing.T) {
		svc, repo := newTestService(t)
		repo.EXPECT().GetByID(gomock.Any(), id).Return(nil, domain.ErrDeliveryNotFound)

		got, err := svc.GetByID(context.Background(), id)
		require.Nil(t, got)
		require.ErrorIs(t, err, domain.ErrDeliveryNotFound)
	})
}

func Test_Service_GetFromEvent(t *testing.T) {
	deliveries := []domain.Delivery{
		{ID: deliveryID, EventID: eventID, Status: domain.StatusPending},
		{ID: deliveryID2, EventID: eventID, Status: domain.StatusFailed},
	}

	t.Run("success", func(t *testing.T) {
		svc, repo := newTestService(t)
		repo.EXPECT().GetFromEvent(gomock.Any(), dto.GetDeliveriesFromEventCommand{
			EventID: eventID,
		}).Return(deliveries, 12, nil)

		got, total, err := svc.GetFromEvent(context.Background(), dto.GetDeliveriesFromEventCommand{
			EventID: eventID,
		})

		require.NoError(t, err)
		require.Equal(t, deliveries, got)
		require.Equal(t, 12, total)
	})

	t.Run("error is wrapped", func(t *testing.T) {
		svc, repo := newTestService(t)
		repo.EXPECT().GetFromEvent(gomock.Any(), dto.GetDeliveriesFromEventCommand{
			EventID: eventID,
		}).Return(nil, 0, domain.ErrEventNotFound)

		got, total, err := svc.GetFromEvent(context.Background(), dto.GetDeliveriesFromEventCommand{
			EventID: eventID,
		})

		require.Nil(t, got)
		require.Zero(t, total)
		require.ErrorIs(t, err, domain.ErrEventNotFound)
	})
}

func Test_Service_Retry(t *testing.T) {
	id := deliveryID

	t.Run("success resets the delivery to pending", func(t *testing.T) {
		svc, repo := newTestService(t)

		var got dto.UpdateDeliveryStatusCommand
		repo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, cmd dto.UpdateDeliveryStatusCommand) error {
				got = cmd
				return nil
			})

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
		svc, repo := newTestService(t)
		repo.EXPECT().UpdateStatus(gomock.Any(), gomock.Any()).Return(domain.ErrDeliveryNotFound)

		require.ErrorIs(t, svc.Retry(context.Background(), id), domain.ErrDeliveryNotFound)
	})
}
