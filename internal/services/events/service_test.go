package events

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
)

func newTestService(t *testing.T) (*Service, *MockEventRepo, *MockDeliveryRepo, *MockEventNotifier) {
	ctrl := gomock.NewController(t)
	eventRepo := NewMockEventRepo(ctrl)
	deliveryRepo := NewMockDeliveryRepo(ctrl)
	notifier := NewMockEventNotifier(ctrl)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewService(eventRepo, deliveryRepo, notifier, log), eventRepo, deliveryRepo, notifier
}

func Test_Service_Publish(t *testing.T) {
	cmd := dto.PublishEventCommand{Type: "order.created"}
	event := &domain.Event{ID: "ev-1", Type: "order.created"}

	t.Run("success", func(t *testing.T) {
		svc, eventRepo, deliveryRepo, notifier := newTestService(t)
		eventRepo.EXPECT().Add(gomock.Any(), cmd).Return(event, nil)
		deliveryRepo.EXPECT().CreateForEvent(gomock.Any(), "ev-1", "order.created").Return(3, nil)
		notifier.EXPECT().Notify()

		got, err := svc.Publish(context.Background(), cmd)
		require.NoError(t, err)
		require.Equal(t, &dto.PublishEventResult{Event: *event, DeliveriesCreated: 3}, got)
	})

	t.Run("event repo error", func(t *testing.T) {
		svc, eventRepo, _, _ := newTestService(t)
		repoErr := errors.New("db down")
		eventRepo.EXPECT().Add(gomock.Any(), cmd).Return(nil, repoErr)

		got, err := svc.Publish(context.Background(), cmd)
		require.Nil(t, got)
		require.ErrorIs(t, err, repoErr)
	})

	t.Run("delivery repo error", func(t *testing.T) {
		svc, eventRepo, deliveryRepo, _ := newTestService(t)
		deliErr := errors.New("redis down")
		eventRepo.EXPECT().Add(gomock.Any(), cmd).Return(event, nil)
		deliveryRepo.EXPECT().CreateForEvent(gomock.Any(), "ev-1", "order.created").Return(0, deliErr)

		got, err := svc.Publish(context.Background(), cmd)
		require.Nil(t, got)
		require.ErrorIs(t, err, deliErr)
	})
}

func Test_Service_Get(t *testing.T) {
	const id = "ev-1"
	event := &domain.Event{ID: id, Type: "order.created"}

	t.Run("success", func(t *testing.T) {
		svc, eventRepo, _, _ := newTestService(t)
		eventRepo.EXPECT().GetByID(gomock.Any(), id).Return(event, nil)

		got, err := svc.Get(context.Background(), id)
		require.NoError(t, err)
		require.Equal(t, &dto.GetEventResult{Event: *event}, got)
	})

	t.Run("error is wrapped", func(t *testing.T) {
		svc, eventRepo, _, _ := newTestService(t)
		eventRepo.EXPECT().GetByID(gomock.Any(), id).Return(nil, domain.ErrEventNotFound)

		got, err := svc.Get(context.Background(), id)
		require.Nil(t, got)
		require.ErrorIs(t, err, domain.ErrEventNotFound)
	})
}
