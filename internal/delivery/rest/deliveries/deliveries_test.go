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

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	now        = time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	errService = errors.New("db down")

	deliveryID = uuid.MustParse("01900000-0000-7000-8000-000000000001")
	endpointID = uuid.MustParse("01900000-0000-7000-8000-0000000000a1")
	eventID    = uuid.MustParse("01900000-0000-7000-8000-0000000000e1")

	publicDeliveryID = domain.ToPublicDeliveryID(deliveryID)
	publicEndpointID = domain.ToPublicEndpointID(endpointID)
	publicEventID    = domain.ToPublicEventID(eventID)
)

func Test_GetDelivery(t *testing.T) {
	delivery := &domain.Delivery{
		ID:               deliveryID,
		EndpointID:       endpointID,
		EventID:          eventID,
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
			request: apiv1.GetDeliveryRequestObject{Id: publicDeliveryID},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetByID(gomock.Any(), deliveryID).Return(delivery, nil)
			},
			want: apiv1.GetDelivery200JSONResponse{
				Delivery: apiv1.DeliveryInfo{
					Id:               publicDeliveryID,
					EndpointId:       publicEndpointID,
					EventId:          publicEventID,
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
			request: apiv1.GetDeliveryRequestObject{Id: publicDeliveryID},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetByID(gomock.Any(), deliveryID).
					Return(nil, fmt.Errorf("service: %w", domain.ErrDeliveryNotFound))
			},
			want: apiv1.GetDelivery404JSONResponse(errDeliveryNotFound),
		},
		{
			name:    "context canceled",
			request: apiv1.GetDeliveryRequestObject{Id: publicDeliveryID},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetByID(gomock.Any(), deliveryID).Return(nil, context.Canceled)
			},
			want: apiv1.GetDelivery500JSONResponse(common.InternalError()),
		},
		{
			name:    "internal error",
			request: apiv1.GetDeliveryRequestObject{Id: publicDeliveryID},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetByID(gomock.Any(), deliveryID).Return(nil, errService)
			},
			want: apiv1.GetDelivery500JSONResponse(common.InternalError()),
		},
		{
			name:    "id with wrong prefix",
			request: apiv1.GetDeliveryRequestObject{Id: domain.ToPublicEndpointID(deliveryID)},
			want:    apiv1.GetDelivery404JSONResponse(errDeliveryNotFound),
		},
		{
			name:    "malformed uuid",
			request: apiv1.GetDeliveryRequestObject{Id: "del_123"},
			want:    apiv1.GetDelivery404JSONResponse(errDeliveryNotFound),
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
		ID:          deliveryID,
		EndpointID:  endpointID,
		EventID:     eventID,
		Status:      domain.StatusPending,
		MaxAttempts: 5,
		NextRetryAt: now,
		CreatedAt:   now,
	}

	deliveryInfo := apiv1.DeliveryInfo{
		Id:          publicDeliveryID,
		EndpointId:  publicEndpointID,
		EventId:     publicEventID,
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
				Id:     publicEventID,
				Params: apiv1.GetEventDeliveriesParams{Page: new(2), Limit: new(20)},
			},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetFromEvent(gomock.Any(), dto.GetDeliveriesFromEventCommand{EventID: eventID, Page: 2, Limit: 20}).
					Return([]domain.Delivery{delivery}, 21, nil)
			},
			want: apiv1.GetEventDeliveries200JSONResponse{
				Deliveries: []apiv1.DeliveryInfo{deliveryInfo},
				Total:      21,
			},
		},
		{
			name:    "default pagination params",
			request: apiv1.GetEventDeliveriesRequestObject{Id: publicEventID},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetFromEvent(gomock.Any(), dto.GetDeliveriesFromEventCommand{EventID: eventID, Page: 1, Limit: 10}).
					Return([]domain.Delivery{delivery}, 1, nil)
			},
			want: apiv1.GetEventDeliveries200JSONResponse{
				Deliveries: []apiv1.DeliveryInfo{deliveryInfo},
				Total:      1,
			},
		},
		{
			name:    "no deliveries returns empty array",
			request: apiv1.GetEventDeliveriesRequestObject{Id: publicEventID},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetFromEvent(gomock.Any(), gomock.Any()).Return(nil, 0, nil)
			},
			want: apiv1.GetEventDeliveries200JSONResponse{
				Deliveries: []apiv1.DeliveryInfo{},
				Total:      0,
			},
		},
		{
			name:    "event not found",
			request: apiv1.GetEventDeliveriesRequestObject{Id: publicEventID},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetFromEvent(gomock.Any(), gomock.Any()).
					Return(nil, 0, fmt.Errorf("service: %w", domain.ErrEventNotFound))
			},
			want: apiv1.GetEventDeliveries404JSONResponse(errEventNotFound),
		},
		{
			name:    "internal error",
			request: apiv1.GetEventDeliveriesRequestObject{Id: publicEventID},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetFromEvent(gomock.Any(), gomock.Any()).Return(nil, 0, errService)
			},
			want: apiv1.GetEventDeliveries500JSONResponse(common.InternalError()),
		},
		{
			name:    "invalid event id",
			request: apiv1.GetEventDeliveriesRequestObject{Id: "ev_not-a-uuid"},
			want:    apiv1.GetEventDeliveries404JSONResponse(errEventNotFound),
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
			request: apiv1.RetryDeliveryRequestObject{Id: publicDeliveryID},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Retry(gomock.Any(), deliveryID).Return(nil)
			},
			want: apiv1.RetryDelivery202Response{},
		},
		{
			name:    "delivery not found",
			request: apiv1.RetryDeliveryRequestObject{Id: publicDeliveryID},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Retry(gomock.Any(), deliveryID).
					Return(fmt.Errorf("service: %w", domain.ErrDeliveryNotFound))
			},
			want: apiv1.RetryDelivery404JSONResponse(errDeliveryNotFound),
		},
		{
			name:    "internal error",
			request: apiv1.RetryDeliveryRequestObject{Id: publicDeliveryID},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Retry(gomock.Any(), deliveryID).Return(errService)
			},
			want: apiv1.RetryDelivery500JSONResponse(common.InternalError()),
		},
		{
			name:    "invalid id",
			request: apiv1.RetryDeliveryRequestObject{Id: "del_123"},
			want:    apiv1.RetryDelivery404JSONResponse(errDeliveryNotFound),
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
