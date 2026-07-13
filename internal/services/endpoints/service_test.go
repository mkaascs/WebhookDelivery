package endpoints

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
	"webhook-delivery/internal/mocks"
)

func newTestService(t *testing.T) (*Service, *mocks.MockRepo) {
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockRepo(ctrl)
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewService(log, repo), repo
}

func Test_Service_Register(t *testing.T) {
	cmd := dto.RegisterEndpointCommand{URL: "https://hooks.example.com", EventTypes: []string{"order.created"}}
	want := &dto.RegisterEndpointResult{ID: "ep-1", Secret: "whsec_x", IsActive: true}

	t.Run("success", func(t *testing.T) {
		svc, repo := newTestService(t)
		repo.EXPECT().AddEndpoint(gomock.Any(), cmd).Return(want, nil)

		got, err := svc.Register(context.Background(), cmd)
		require.NoError(t, err)
		require.Equal(t, want, got)
	})

	t.Run("error is wrapped", func(t *testing.T) {
		svc, repo := newTestService(t)
		repo.EXPECT().AddEndpoint(gomock.Any(), cmd).Return(nil, domain.ErrEndpointNotFound)

		got, err := svc.Register(context.Background(), cmd)
		require.Nil(t, got)
		require.ErrorIs(t, err, domain.ErrEndpointNotFound)
	})
}

func Test_Service_GetByID(t *testing.T) {
	const id = "ep-1"
	want := &dto.GetEndpointResult{ID: id, URL: "https://hooks.example.com"}

	t.Run("success", func(t *testing.T) {
		svc, repo := newTestService(t)
		repo.EXPECT().GetByID(gomock.Any(), id).Return(&domain.Endpoint{
			ID:  want.ID,
			URL: want.URL,
		}, nil)

		got, err := svc.GetByID(context.Background(), id)
		require.NoError(t, err)
		require.Equal(t, want, got)
	})

	t.Run("error is wrapped", func(t *testing.T) {
		svc, repo := newTestService(t)
		repo.EXPECT().GetByID(gomock.Any(), id).Return(nil, domain.ErrEndpointNotFound)

		got, err := svc.GetByID(context.Background(), id)
		require.Nil(t, got)
		require.ErrorIs(t, err, domain.ErrEndpointNotFound)
	})
}

func Test_Service_GetAll(t *testing.T) {
	cmd := dto.GetAllEndpointsCommand{Page: 1, Limit: 10}
	want := &dto.GetAllEndpointsResult{Total: 1, Endpoints: []dto.GetEndpointResult{{ID: "ep-1"}}}

	t.Run("success", func(t *testing.T) {
		svc, repo := newTestService(t)
		repo.EXPECT().GetAll(gomock.Any(), cmd).Return([]domain.Endpoint{{ID: "ep-1"}}, 1, nil)

		got, err := svc.GetAll(context.Background(), cmd)
		require.NoError(t, err)
		require.Equal(t, want, got)
	})

	t.Run("error is wrapped", func(t *testing.T) {
		svc, repo := newTestService(t)
		repo.EXPECT().GetAll(gomock.Any(), cmd).Return(nil, 0, domain.ErrEndpointNotFound)

		got, err := svc.GetAll(context.Background(), cmd)
		require.Nil(t, got)
		require.ErrorIs(t, err, domain.ErrEndpointNotFound)
	})
}

func Test_Service_Update(t *testing.T) {
	cmd := dto.UpdateEndpointCommand{ID: "ep-1"}

	t.Run("success", func(t *testing.T) {
		svc, repo := newTestService(t)
		repo.EXPECT().Update(gomock.Any(), cmd).Return(nil)

		require.NoError(t, svc.Update(context.Background(), cmd))
	})

	t.Run("error is wrapped", func(t *testing.T) {
		svc, repo := newTestService(t)
		repo.EXPECT().Update(gomock.Any(), cmd).Return(domain.ErrEndpointNotFound)

		require.ErrorIs(t, svc.Update(context.Background(), cmd), domain.ErrEndpointNotFound)
	})
}

func Test_Service_Delete(t *testing.T) {
	const id = "ep-1"

	t.Run("success", func(t *testing.T) {
		svc, repo := newTestService(t)
		repo.EXPECT().Delete(gomock.Any(), id).Return(nil)

		require.NoError(t, svc.Delete(context.Background(), id))
	})

	t.Run("error is wrapped", func(t *testing.T) {
		svc, repo := newTestService(t)
		repo.EXPECT().Delete(gomock.Any(), id).Return(domain.ErrEndpointNotFound)

		require.ErrorIs(t, svc.Delete(context.Background(), id), domain.ErrEndpointNotFound)
	})
}
