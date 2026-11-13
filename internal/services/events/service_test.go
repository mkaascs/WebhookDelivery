package events

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
)

var eventID = uuid.MustParse("01900000-0000-7000-8000-0000000000e1")

func newTestService(t *testing.T) (*Service, *MockEventRepo, *MockDeliveryRepo) {
	ctrl := gomock.NewController(t)
	eventRepo := NewMockEventRepo(ctrl)
	deliveryRepo := NewMockDeliveryRepo(ctrl)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewService(eventRepo, deliveryRepo, log), eventRepo, deliveryRepo
}

func Test_Service_Publish(t *testing.T) {
	cmd := dto.PublishEventCommand{Type: "order.created"}
	event := &domain.Event{ID: eventID, Type: "order.created"}

	t.Run("success", func(t *testing.T) {
		svc, eventRepo, deliveryRepo := newTestService(t)
		eventRepo.EXPECT().Add(gomock.Any(), cmd).Return(event, nil)
		deliveryRepo.EXPECT().CreateForEvent(gomock.Any(), eventID, "order.created").Return(3, nil)

		got, deliveriesCreated, err := svc.Publish(context.Background(), cmd)
		require.NoError(t, err)
		require.Equal(t, event, got)
		require.Equal(t, 3, deliveriesCreated)
	})

	t.Run("event repo error", func(t *testing.T) {
		svc, eventRepo, _ := newTestService(t)
		repoErr := errors.New("db down")
		eventRepo.EXPECT().Add(gomock.Any(), cmd).Return(nil, repoErr)

		got, _, err := svc.Publish(context.Background(), cmd)
		require.Nil(t, got)
		require.ErrorIs(t, err, repoErr)
	})

	t.Run("delivery repo error", func(t *testing.T) {
		svc, eventRepo, deliveryRepo := newTestService(t)
		deliErr := errors.New("redis down")
		eventRepo.EXPECT().Add(gomock.Any(), cmd).Return(event, nil)
		deliveryRepo.EXPECT().CreateForEvent(gomock.Any(), eventID, "order.created").Return(0, deliErr)

		got, _, err := svc.Publish(context.Background(), cmd)
		require.Nil(t, got)
		require.ErrorIs(t, err, deliErr)
	})
}

func Test_Service_Get(t *testing.T) {
	id := eventID
	event := &domain.Event{ID: id, Type: "order.created"}

	t.Run("success", func(t *testing.T) {
		svc, eventRepo, _ := newTestService(t)
		eventRepo.EXPECT().GetByID(gomock.Any(), id).Return(event, nil)

		got, err := svc.Get(context.Background(), id)
		require.NoError(t, err)
		require.Equal(t, event, got)
	})

	t.Run("error is wrapped", func(t *testing.T) {
		svc, eventRepo, _ := newTestService(t)
		eventRepo.EXPECT().GetByID(gomock.Any(), id).Return(nil, domain.ErrEventNotFound)

		got, err := svc.Get(context.Background(), id)
		require.Nil(t, got)
		require.ErrorIs(t, err, domain.ErrEventNotFound)
	})
}
