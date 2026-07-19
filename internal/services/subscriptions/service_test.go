package subscriptions

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
)

func newTestService(t *testing.T) (*Service, *MockSubscriptionRepo) {
	ctrl := gomock.NewController(t)
	repo := NewMockSubscriptionRepo(ctrl)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewService(repo, log), repo
}

func Test_Service_Add(t *testing.T) {
	cmd := dto.AddSubscriptionCommand{EndpointID: "ep-1", EventTypes: []string{"order.created"}}
	subs := []domain.Subscription{{ID: "sub-1", EndpointID: "ep-1", EventType: "order.created"}}

	t.Run("success", func(t *testing.T) {
		svc, repo := newTestService(t)
		repo.EXPECT().Add(gomock.Any(), cmd).Return(subs, nil)

		got, err := svc.Add(context.Background(), cmd)
		require.NoError(t, err)
		require.Equal(t, &dto.AddSubscriptionResult{Subscriptions: subs}, got)
	})

	t.Run("error is wrapped", func(t *testing.T) {
		svc, repo := newTestService(t)
		repo.EXPECT().Add(gomock.Any(), cmd).Return(nil, domain.ErrSubscriptionAlreadyExists)

		got, err := svc.Add(context.Background(), cmd)
		require.Nil(t, got)
		require.ErrorIs(t, err, domain.ErrSubscriptionAlreadyExists)
	})
}

func Test_Service_GetAll(t *testing.T) {
	const endpointID = "ep-1"
	subs := []domain.Subscription{{ID: "sub-1", EndpointID: endpointID, EventType: "order.created"}}
	want := []dto.GetSubscriptionResult{{ID: "sub-1", EndpointID: endpointID, EventType: "order.created"}}

	t.Run("success", func(t *testing.T) {
		svc, repo := newTestService(t)
		repo.EXPECT().GetAll(gomock.Any(), endpointID).Return(subs, nil)

		got, err := svc.GetAll(context.Background(), endpointID)
		require.NoError(t, err)
		require.Equal(t, want, got)
	})

	t.Run("error is wrapped", func(t *testing.T) {
		svc, repo := newTestService(t)
		repo.EXPECT().GetAll(gomock.Any(), endpointID).Return(nil, domain.ErrEndpointNotFound)

		got, err := svc.GetAll(context.Background(), endpointID)
		require.Nil(t, got)
		require.ErrorIs(t, err, domain.ErrEndpointNotFound)
	})
}

func Test_Service_Delete(t *testing.T) {
	const id = "sub-1"

	t.Run("success", func(t *testing.T) {
		svc, repo := newTestService(t)
		repo.EXPECT().Delete(gomock.Any(), id).Return(nil)

		require.NoError(t, svc.Delete(context.Background(), id))
	})

	t.Run("error is wrapped", func(t *testing.T) {
		svc, repo := newTestService(t)
		repo.EXPECT().Delete(gomock.Any(), id).Return(domain.ErrSubscriptionNotFound)

		require.ErrorIs(t, svc.Delete(context.Background(), id), domain.ErrSubscriptionNotFound)
	})
}
