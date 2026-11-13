package endpoints

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
	apiv1 "webhook-delivery/gen/rest/v1"
	"webhook-delivery/internal/delivery/rest/common"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	now        = time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	errService = errors.New("db down")

	endpointID       = uuid.MustParse("01900000-0000-7000-8000-0000000000a1")
	publicEndpointID = domain.ToPublicEndpointID(endpointID)
)

func Test_GetAllEndpoints(t *testing.T) {
	endpoint := domain.Endpoint{
		ID:          endpointID,
		URL:         "https://example.com/webhooks",
		EventTypes:  []string{"order.created"},
		Description: new("Billing webhook"),
		IsActive:    true,
		CreatedAt:   now,
	}

	endpointInfo := apiv1.EndpointInfo{
		Id:          publicEndpointID,
		Url:         "https://example.com/webhooks",
		EventTypes:  []string{"order.created"},
		Description: "Billing webhook",
		IsActive:    true,
		CreatedAt:   now,
	}

	tests := []struct {
		name      string
		request   apiv1.GetAllEndpointsRequestObject
		setupMock func(m *MockInterface)
		want      apiv1.GetAllEndpointsResponseObject
	}{
		{
			name: "success with pagination params",
			request: apiv1.GetAllEndpointsRequestObject{
				Params: apiv1.GetAllEndpointsParams{Page: new(3), Limit: new(50)},
			},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetAll(gomock.Any(), dto.GetAllEndpointsCommand{Page: 3, Limit: 50}).
					Return([]domain.Endpoint{endpoint}, 101, nil)
			},
			want: apiv1.GetAllEndpoints200JSONResponse{
				Endpoints: []apiv1.EndpointInfo{endpointInfo},
				Page:      3,
				Limit:     50,
				Total:     101,
			},
		},
		{
			name:    "default pagination params",
			request: apiv1.GetAllEndpointsRequestObject{},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetAll(gomock.Any(), dto.GetAllEndpointsCommand{Page: 1, Limit: 10}).
					Return([]domain.Endpoint{endpoint}, 1, nil)
			},
			want: apiv1.GetAllEndpoints200JSONResponse{
				Endpoints: []apiv1.EndpointInfo{endpointInfo},
				Page:      1,
				Limit:     10,
				Total:     1,
			},
		},
		{
			name:    "nil event types and description are mapped to defaults",
			request: apiv1.GetAllEndpointsRequestObject{},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetAll(gomock.Any(), gomock.Any()).
					Return([]domain.Endpoint{
						{ID: endpointID, URL: "https://example.com", CreatedAt: now},
					}, 1, nil)
			},
			want: apiv1.GetAllEndpoints200JSONResponse{
				Endpoints: []apiv1.EndpointInfo{
					{Id: publicEndpointID, Url: "https://example.com", EventTypes: []string{}, CreatedAt: now},
				},
				Page:  1,
				Limit: 10,
				Total: 1,
			},
		},
		{
			name:    "no endpoints returns empty array",
			request: apiv1.GetAllEndpointsRequestObject{},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetAll(gomock.Any(), gomock.Any()).Return(nil, 0, nil)
			},
			want: apiv1.GetAllEndpoints200JSONResponse{
				Endpoints: []apiv1.EndpointInfo{},
				Page:      1,
				Limit:     10,
			},
		},
		{
			name:    "internal error",
			request: apiv1.GetAllEndpointsRequestObject{},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetAll(gomock.Any(), gomock.Any()).Return(nil, 0, errService)
			},
			want: apiv1.GetAllEndpoints500JSONResponse(common.InternalError()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := NewMockInterface(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(service)
			}

			resp, err := NewHandler(service).GetAllEndpoints(context.Background(), tt.request)

			require.NoError(t, err)
			require.Equal(t, tt.want, resp)
		})
	}
}

func Test_GetEndpoint(t *testing.T) {
	tests := []struct {
		name      string
		request   apiv1.GetEndpointRequestObject
		setupMock func(m *MockInterface)
		want      apiv1.GetEndpointResponseObject
	}{
		{
			name:    "success",
			request: apiv1.GetEndpointRequestObject{Id: publicEndpointID},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetByID(gomock.Any(), endpointID).Return(&domain.Endpoint{
					ID:          endpointID,
					URL:         "https://example.com/webhooks",
					EventTypes:  []string{"order.created", "order.paid"},
					Description: new("Billing webhook"),
					IsActive:    true,
					CreatedAt:   now,
				}, nil)
			},
			want: apiv1.GetEndpoint200JSONResponse{
				Id:          publicEndpointID,
				Url:         "https://example.com/webhooks",
				EventTypes:  []string{"order.created", "order.paid"},
				Description: "Billing webhook",
				IsActive:    true,
				CreatedAt:   now,
			},
		},
		{
			name:    "nil event types and description are mapped to defaults",
			request: apiv1.GetEndpointRequestObject{Id: publicEndpointID},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetByID(gomock.Any(), endpointID).Return(&domain.Endpoint{
					ID:        endpointID,
					URL:       "https://example.com/webhooks",
					CreatedAt: now,
				}, nil)
			},
			want: apiv1.GetEndpoint200JSONResponse{
				Id:         publicEndpointID,
				Url:        "https://example.com/webhooks",
				EventTypes: []string{},
				CreatedAt:  now,
			},
		},
		{
			name:    "endpoint not found",
			request: apiv1.GetEndpointRequestObject{Id: publicEndpointID},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetByID(gomock.Any(), endpointID).
					Return(nil, fmt.Errorf("service: %w", domain.ErrEndpointNotFound))
			},
			want: apiv1.GetEndpoint404JSONResponse(errEndpointNotFound),
		},
		{
			name:    "internal error",
			request: apiv1.GetEndpointRequestObject{Id: publicEndpointID},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetByID(gomock.Any(), endpointID).Return(nil, errService)
			},
			want: apiv1.GetEndpoint500JSONResponse(common.InternalError()),
		},
		{
			name:    "invalid id",
			request: apiv1.GetEndpointRequestObject{Id: "ep_123"},
			want:    apiv1.GetEndpoint404JSONResponse(errEndpointNotFound),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := NewMockInterface(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(service)
			}

			resp, err := NewHandler(service).GetEndpoint(context.Background(), tt.request)

			require.NoError(t, err)
			require.Equal(t, tt.want, resp)
		})
	}
}

func Test_RegisterEndpoint(t *testing.T) {
	const url = "https://example.com/webhooks"

	tests := []struct {
		name      string
		request   apiv1.RegisterEndpointRequestObject
		setupMock func(m *MockInterface)
		want      apiv1.RegisterEndpointResponseObject
	}{
		{
			name: "success",
			request: apiv1.RegisterEndpointRequestObject{Body: &apiv1.RegisterEndpointRequest{
				Url:         url,
				EventTypes:  &[]string{"order.created"},
				Description: new("Billing webhook"),
			}},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Register(gomock.Any(), dto.RegisterEndpointCommand{
					URL:         url,
					EventTypes:  []string{"order.created"},
					Description: new("Billing webhook"),
				}).Return(&domain.Endpoint{
					ID:          endpointID,
					URL:         url,
					Secret:      "whsec_123",
					EventTypes:  []string{"order.created"},
					Description: new("Billing webhook"),
					IsActive:    true,
					CreatedAt:   now,
				}, nil)
			},
			want: apiv1.RegisterEndpoint200JSONResponse{
				Id:         publicEndpointID,
				Url:        url,
				Secret:     "whsec_123",
				IsActive:   true,
				EventTypes: []string{"order.created"},
				CreatedAt:  now,
			},
		},
		{
			name:    "without event types",
			request: apiv1.RegisterEndpointRequestObject{Body: &apiv1.RegisterEndpointRequest{Url: url}},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Register(gomock.Any(), dto.RegisterEndpointCommand{
					URL:        url,
					EventTypes: []string{},
				}).Return(&domain.Endpoint{
					ID:         endpointID,
					URL:        url,
					Secret:     "whsec_123",
					EventTypes: []string{},
					IsActive:   true,
					CreatedAt:  now,
				}, nil)
			},
			want: apiv1.RegisterEndpoint200JSONResponse{
				Id:         publicEndpointID,
				Url:        url,
				Secret:     "whsec_123",
				IsActive:   true,
				EventTypes: []string{},
				CreatedAt:  now,
			},
		},
		{
			name:    "invalid url scheme",
			request: apiv1.RegisterEndpointRequestObject{Body: &apiv1.RegisterEndpointRequest{Url: "ftp://example.com"}},
			want:    apiv1.RegisterEndpoint400JSONResponse(common.InvalidURLError()),
		},
		{
			name:    "url without host",
			request: apiv1.RegisterEndpointRequestObject{Body: &apiv1.RegisterEndpointRequest{Url: "not a url"}},
			want:    apiv1.RegisterEndpoint400JSONResponse(common.InvalidURLError()),
		},
		{
			name:    "localhost url",
			request: apiv1.RegisterEndpointRequestObject{Body: &apiv1.RegisterEndpointRequest{Url: "http://localhost:8080/hook"}},
			want:    apiv1.RegisterEndpoint400JSONResponse(common.InvalidURLError()),
		},
		{
			name:    "loopback ip url",
			request: apiv1.RegisterEndpointRequestObject{Body: &apiv1.RegisterEndpointRequest{Url: "http://127.0.0.1/hook"}},
			want:    apiv1.RegisterEndpoint400JSONResponse(common.InvalidURLError()),
		},
		{
			name:    "link-local ip url",
			request: apiv1.RegisterEndpointRequestObject{Body: &apiv1.RegisterEndpointRequest{Url: "http://169.254.169.254/latest"}},
			want:    apiv1.RegisterEndpoint400JSONResponse(common.InvalidURLError()),
		},
		{
			name:    "internal error",
			request: apiv1.RegisterEndpointRequestObject{Body: &apiv1.RegisterEndpointRequest{Url: url}},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Register(gomock.Any(), gomock.Any()).Return(nil, errService)
			},
			want: apiv1.RegisterEndpoint500JSONResponse(common.InternalError()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := NewMockInterface(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(service)
			}

			resp, err := NewHandler(service).RegisterEndpoint(context.Background(), tt.request)

			require.NoError(t, err)
			require.Equal(t, tt.want, resp)
		})
	}
}

func Test_UpdateEndpoint(t *testing.T) {
	tests := []struct {
		name      string
		request   apiv1.UpdateEndpointRequestObject
		setupMock func(m *MockInterface)
		want      apiv1.UpdateEndpointResponseObject
	}{
		{
			name: "success",
			request: apiv1.UpdateEndpointRequestObject{
				Id: publicEndpointID,
				Body: &apiv1.UpdateEndpointRequest{
					Url:         new("https://example.com/new"),
					IsActive:    new(false),
					Description: new("new description"),
				},
			},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Update(gomock.Any(), dto.UpdateEndpointCommand{
					ID:          endpointID,
					URL:         new("https://example.com/new"),
					IsActive:    new(false),
					Description: new("new description"),
				}).Return(nil)
			},
			want: apiv1.UpdateEndpoint204Response{},
		},
		{
			name: "success without url",
			request: apiv1.UpdateEndpointRequestObject{
				Id:   publicEndpointID,
				Body: &apiv1.UpdateEndpointRequest{IsActive: new(true)},
			},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Update(gomock.Any(), dto.UpdateEndpointCommand{
					ID:       endpointID,
					IsActive: new(true),
				}).Return(nil)
			},
			want: apiv1.UpdateEndpoint204Response{},
		},
		{
			name:    "empty body is a no-op",
			request: apiv1.UpdateEndpointRequestObject{Id: publicEndpointID},
			want:    apiv1.UpdateEndpoint204Response{},
		},
		{
			name: "invalid url",
			request: apiv1.UpdateEndpointRequestObject{
				Id:   publicEndpointID,
				Body: &apiv1.UpdateEndpointRequest{Url: new("http://localhost/hook")},
			},
			want: apiv1.UpdateEndpoint400JSONResponse(common.InvalidURLError()),
		},
		{
			name: "endpoint not found",
			request: apiv1.UpdateEndpointRequestObject{
				Id:   publicEndpointID,
				Body: &apiv1.UpdateEndpointRequest{IsActive: new(true)},
			},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Update(gomock.Any(), gomock.Any()).
					Return(fmt.Errorf("service: %w", domain.ErrEndpointNotFound))
			},
			want: apiv1.UpdateEndpoint404JSONResponse(errEndpointNotFound),
		},
		{
			name: "internal error",
			request: apiv1.UpdateEndpointRequestObject{
				Id:   publicEndpointID,
				Body: &apiv1.UpdateEndpointRequest{IsActive: new(true)},
			},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errService)
			},
			want: apiv1.UpdateEndpoint500JSONResponse(common.InternalError()),
		},
		{
			name: "invalid id",
			request: apiv1.UpdateEndpointRequestObject{
				Id:   "sub_" + endpointID.String(),
				Body: &apiv1.UpdateEndpointRequest{IsActive: new(true)},
			},
			want: apiv1.UpdateEndpoint404JSONResponse(errEndpointNotFound),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := NewMockInterface(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(service)
			}

			resp, err := NewHandler(service).UpdateEndpoint(context.Background(), tt.request)

			require.NoError(t, err)
			require.Equal(t, tt.want, resp)
		})
	}
}

func Test_DeleteEndpoint(t *testing.T) {
	tests := []struct {
		name      string
		request   apiv1.DeleteEndpointRequestObject
		setupMock func(m *MockInterface)
		want      apiv1.DeleteEndpointResponseObject
	}{
		{
			name:    "success",
			request: apiv1.DeleteEndpointRequestObject{Id: publicEndpointID},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Delete(gomock.Any(), endpointID).Return(nil)
			},
			want: apiv1.DeleteEndpoint204Response{},
		},
		{
			name:    "endpoint not found",
			request: apiv1.DeleteEndpointRequestObject{Id: publicEndpointID},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Delete(gomock.Any(), endpointID).
					Return(fmt.Errorf("service: %w", domain.ErrEndpointNotFound))
			},
			want: apiv1.DeleteEndpoint404JSONResponse(errEndpointNotFound),
		},
		{
			name:    "internal error",
			request: apiv1.DeleteEndpointRequestObject{Id: publicEndpointID},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Delete(gomock.Any(), endpointID).Return(errService)
			},
			want: apiv1.DeleteEndpoint500JSONResponse(common.InternalError()),
		},
		{
			name:    "invalid id",
			request: apiv1.DeleteEndpointRequestObject{Id: "ep_"},
			want:    apiv1.DeleteEndpoint404JSONResponse(errEndpointNotFound),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := NewMockInterface(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(service)
			}

			resp, err := NewHandler(service).DeleteEndpoint(context.Background(), tt.request)

			require.NoError(t, err)
			require.Equal(t, tt.want, resp)
		})
	}
}
