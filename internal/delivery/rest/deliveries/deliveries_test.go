package deliveries

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

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	now        = time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	errService = errors.New("db down")
)

func Test_GetDelivery(t *testing.T) {
	delivery := domain.Delivery{
		ID:               "del-1",
		EndpointID:       "ep-1",
		EventID:          "ev-1",
		Status:           domain.StatusFailed,
		Attempts:         2,
		MaxAttempts:      5,
		NextRetryAt:      now.Add(time.Minute),
		CreatedAt:        now,
		LastResponseCode: new(500),
		LastError:        new("HTTP 500"),
	}

	tests := []struct {
		name      string
		request   apiv1.GetDeliveryRequestObject
		setupMock func(m *MockInterface)
		want      apiv1.GetDeliveryResponseObject
	}{
		{
			name:    "success",
			request: apiv1.GetDeliveryRequestObject{Id: "del-1"},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetByID(gomock.Any(), "del-1").
					Return(&dto.GetDeliveryResult{Delivery: delivery}, nil)
			},

			want: apiv1.GetDelivery200JSONResponse{
				Delivery: apiv1.DeliveryInfo{
					Id:               "del-1",
					EndpointId:       "ep-1",
					EventId:          "ev-1",
					Status:           apiv1.Failed,
					Attempts:         2,
					MaxAttempts:      5,
					NextRetryAt:      now.Add(time.Minute),
					CreatedAt:        now,
					LastResponseCode: new(500),
					LastError:        new("HTTP 500"),
				},
			},
		},
		{
			name:    "delivery not found",
			request: apiv1.GetDeliveryRequestObject{Id: "del-404"},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetByID(gomock.Any(), "del-404").
					Return(nil, fmt.Errorf("service: %w", domain.ErrDeliveryNotFound))
			},
			want: apiv1.GetDelivery404JSONResponse(errDeliveryNotFound),
		},
		{
			name:    "context canceled",
			request: apiv1.GetDeliveryRequestObject{Id: "del-1"},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetByID(gomock.Any(), "del-1").Return(nil, context.Canceled)
			},
			want: apiv1.GetDelivery500JSONResponse(common.InternalError()),
		},
		{
			name:    "internal error",
			request: apiv1.GetDeliveryRequestObject{Id: "del-1"},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetByID(gomock.Any(), "del-1").Return(nil, errService)
			},
			want: apiv1.GetDelivery500JSONResponse(common.InternalError()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := NewMockInterface(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(service)
			}

			resp, err := NewHandler(service).GetDelivery(context.Background(), tt.request)

			require.NoError(t, err)
			require.Equal(t, tt.want, resp)
		})
	}
}

func Test_GetEventDeliveries(t *testing.T) {
	delivery := domain.Delivery{
		ID:          "del-1",
		EndpointID:  "ep-1",
		EventID:     "ev-1",
		Status:      domain.StatusPending,
		MaxAttempts: 5,
		NextRetryAt: now,
		CreatedAt:   now,
	}

	deliveryInfo := apiv1.DeliveryInfo{
		Id:          "del-1",
		EndpointId:  "ep-1",
		EventId:     "ev-1",
		Status:      apiv1.Pending,
		MaxAttempts: 5,
		NextRetryAt: now,
		CreatedAt:   now,
	}

	tests := []struct {
		name      string
		request   apiv1.GetEventDeliveriesRequestObject
		setupMock func(m *MockInterface)
		want      apiv1.GetEventDeliveriesResponseObject
	}{
		{
			name: "success with pagination params",
			request: apiv1.GetEventDeliveriesRequestObject{
				Id:     "ev-1",
				Params: apiv1.GetEventDeliveriesParams{Page: new(2), Limit: new(20)},
			},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetFromEvent(gomock.Any(), dto.GetDeliveriesFromEventCommand{EventID: "ev-1", Page: 2, Limit: 20}).
					Return(&dto.GetDeliveriesFromEventResult{Total: 21, Deliveries: []domain.Delivery{delivery}}, nil)
			},
			want: apiv1.GetEventDeliveries200JSONResponse{
				Deliveries: []apiv1.DeliveryInfo{deliveryInfo},
				Total:      21,
			},
		},
		{
			name:    "default pagination params",
			request: apiv1.GetEventDeliveriesRequestObject{Id: "ev-1"},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetFromEvent(gomock.Any(), dto.GetDeliveriesFromEventCommand{EventID: "ev-1", Page: 1, Limit: 10}).
					Return(&dto.GetDeliveriesFromEventResult{Total: 1, Deliveries: []domain.Delivery{delivery}}, nil)
			},
			want: apiv1.GetEventDeliveries200JSONResponse{
				Deliveries: []apiv1.DeliveryInfo{deliveryInfo},
				Total:      1,
			},
		},
		{
			name:    "no deliveries returns empty array",
			request: apiv1.GetEventDeliveriesRequestObject{Id: "ev-1"},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetFromEvent(gomock.Any(), gomock.Any()).
					Return(&dto.GetDeliveriesFromEventResult{}, nil)
			},
			want: apiv1.GetEventDeliveries200JSONResponse{
				Deliveries: []apiv1.DeliveryInfo{},
				Total:      0,
			},
		},
		{
			name:    "internal error",
			request: apiv1.GetEventDeliveriesRequestObject{Id: "ev-1"},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetFromEvent(gomock.Any(), gomock.Any()).Return(nil, errService)
			},
			want: apiv1.GetEventDeliveries500JSONResponse(common.InternalError()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := NewMockInterface(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(service)
			}

			resp, err := NewHandler(service).GetEventDeliveries(context.Background(), tt.request)

			require.NoError(t, err)
			require.Equal(t, tt.want, resp)
		})
	}
}

func Test_RetryDelivery(t *testing.T) {
	tests := []struct {
		name      string
		request   apiv1.RetryDeliveryRequestObject
		setupMock func(m *MockInterface)
		want      apiv1.RetryDeliveryResponseObject
	}{
		{
			name:    "success",
			request: apiv1.RetryDeliveryRequestObject{Id: "del-1"},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Retry(gomock.Any(), "del-1").Return(nil)
			},
			want: apiv1.RetryDelivery202Response{},
		},
		{
			name:    "delivery not found",
			request: apiv1.RetryDeliveryRequestObject{Id: "del-404"},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Retry(gomock.Any(), "del-404").
					Return(fmt.Errorf("service: %w", domain.ErrDeliveryNotFound))
			},
			want: apiv1.RetryDelivery404JSONResponse(errDeliveryNotFound),
		},
		{
			name:    "internal error",
			request: apiv1.RetryDeliveryRequestObject{Id: "del-1"},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Retry(gomock.Any(), "del-1").Return(errService)
			},
			want: apiv1.RetryDelivery500JSONResponse(common.InternalError()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := NewMockInterface(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(service)
			}

			resp, err := NewHandler(service).RetryDelivery(context.Background(), tt.request)

			require.NoError(t, err)
			require.Equal(t, tt.want, resp)
		})
	}
}
