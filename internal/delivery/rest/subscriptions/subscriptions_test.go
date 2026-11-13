package subscriptions

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

	endpointID = uuid.MustParse("01900000-0000-7000-8000-0000000000a1")
	subID1     = uuid.MustParse("01900000-0000-7000-8000-0000000000b1")
	subID2     = uuid.MustParse("01900000-0000-7000-8000-0000000000b2")

	publicEndpointID = domain.ToPublicEndpointID(endpointID)
	publicSubID1     = domain.ToPublicSubscriptionID(subID1)
	publicSubID2     = domain.ToPublicSubscriptionID(subID2)
)

func Test_GetAllSubscriptions(t *testing.T) {
	tests := []struct {
		name      string
		request   apiv1.GetAllSubscriptionsRequestObject
		setupMock func(m *MockInterface)
		want      apiv1.GetAllSubscriptionsResponseObject
	}{
		{
			name:    "success",
			request: apiv1.GetAllSubscriptionsRequestObject{Id: publicEndpointID},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetAll(gomock.Any(), endpointID).Return([]domain.Subscription{
					{ID: subID1, EndpointID: endpointID, EventType: "order.created", CreatedAt: now},
					{ID: subID2, EndpointID: endpointID, EventType: "order.paid", CreatedAt: now},
				}, nil)
			},
			want: apiv1.GetAllSubscriptions200JSONResponse{Subscriptions: []apiv1.SubscriptionInfo{
				{Id: publicSubID1, EndpointId: publicEndpointID, EventType: "order.created", CreatedAt: now},
				{Id: publicSubID2, EndpointId: publicEndpointID, EventType: "order.paid", CreatedAt: now},
			}},
		},
		{
			name:    "no subscriptions returns empty array",
			request: apiv1.GetAllSubscriptionsRequestObject{Id: publicEndpointID},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetAll(gomock.Any(), endpointID).Return(nil, nil)
			},
			want: apiv1.GetAllSubscriptions200JSONResponse{Subscriptions: []apiv1.SubscriptionInfo{}},
		},
		{
			name:    "endpoint not found",
			request: apiv1.GetAllSubscriptionsRequestObject{Id: publicEndpointID},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetAll(gomock.Any(), endpointID).
					Return(nil, fmt.Errorf("service: %w", domain.ErrEndpointNotFound))
			},
			want: apiv1.GetAllSubscriptions404JSONResponse(errEndpointNotFound),
		},
		{
			name:    "internal error",
			request: apiv1.GetAllSubscriptionsRequestObject{Id: publicEndpointID},
			setupMock: func(m *MockInterface) {
				m.EXPECT().GetAll(gomock.Any(), endpointID).Return(nil, errService)
			},
			want: apiv1.GetAllSubscriptions500JSONResponse(common.InternalError()),
		},
		{
			name:    "invalid endpoint id",
			request: apiv1.GetAllSubscriptionsRequestObject{Id: "ep_123"},
			want:    apiv1.GetAllSubscriptions404JSONResponse(errEndpointNotFound),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := NewMockInterface(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(service)
			}

			resp, err := NewHandler(service).GetAllSubscriptions(context.Background(), tt.request)

			require.NoError(t, err)
			require.Equal(t, tt.want, resp)
		})
	}
}

func Test_Subscribe(t *testing.T) {
	request := apiv1.SubscribeRequestObject{
		Id:   publicEndpointID,
		Body: &apiv1.SubscribeRequest{EventTypes: []string{"order.created", "order.paid"}},
	}

	command := dto.AddSubscriptionCommand{
		EndpointID: endpointID,
		EventTypes: []string{"order.created", "order.paid"},
	}

	tests := []struct {
		name      string
		request   apiv1.SubscribeRequestObject
		setupMock func(m *MockInterface)
		want      apiv1.SubscribeResponseObject
	}{
		{
			name:    "success",
			request: request,
			setupMock: func(m *MockInterface) {
				m.EXPECT().Add(gomock.Any(), command).Return([]domain.Subscription{
					{ID: subID1, EndpointID: endpointID, EventType: "order.created", CreatedAt: now},
					{ID: subID2, EndpointID: endpointID, EventType: "order.paid", CreatedAt: now},
				}, nil)
			},
			want: apiv1.Subscribe200JSONResponse{
				EndpointId: publicEndpointID,
				Subscriptions: []apiv1.SubscriptionInfo{
					{Id: publicSubID1, EndpointId: publicEndpointID, EventType: "order.created", CreatedAt: now},
					{Id: publicSubID2, EndpointId: publicEndpointID, EventType: "order.paid", CreatedAt: now},
				},
			},
		},
		{
			name:    "already subscribed to all event types returns empty array",
			request: request,
			setupMock: func(m *MockInterface) {
				m.EXPECT().Add(gomock.Any(), command).Return(nil, nil)
			},
			want: apiv1.Subscribe200JSONResponse{
				EndpointId:    publicEndpointID,
				Subscriptions: []apiv1.SubscriptionInfo{},
			},
		},
		{
			name:    "endpoint not found",
			request: request,
			setupMock: func(m *MockInterface) {
				m.EXPECT().Add(gomock.Any(), command).
					Return(nil, fmt.Errorf("service: %w", domain.ErrEndpointNotFound))
			},
			want: apiv1.Subscribe404JSONResponse(errEndpointNotFound),
		},
		{
			name:    "subscription already exists",
			request: request,
			setupMock: func(m *MockInterface) {
				m.EXPECT().Add(gomock.Any(), command).
					Return(nil, fmt.Errorf("service: %w", domain.ErrSubscriptionAlreadyExists))
			},
			want: apiv1.Subscribe409JSONResponse(errSubscriptionAlreadyExists),
		},
		{
			name:    "internal error",
			request: request,
			setupMock: func(m *MockInterface) {
				m.EXPECT().Add(gomock.Any(), command).Return(nil, errService)
			},
			want: apiv1.Subscribe500JSONResponse(common.InternalError()),
		},
		{
			name: "invalid endpoint id",
			request: apiv1.SubscribeRequestObject{
				Id:   "endpoint-1",
				Body: &apiv1.SubscribeRequest{EventTypes: []string{"order.created"}},
			},
			want: apiv1.Subscribe404JSONResponse(errEndpointNotFound),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := NewMockInterface(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(service)
			}

			resp, err := NewHandler(service).Subscribe(context.Background(), tt.request)

			require.NoError(t, err)
			require.Equal(t, tt.want, resp)
		})
	}
}

func Test_DeleteSubscription(t *testing.T) {
	tests := []struct {
		name      string
		request   apiv1.DeleteSubscriptionRequestObject
		setupMock func(m *MockInterface)
		want      apiv1.DeleteSubscriptionResponseObject
	}{
		{
			name:    "success",
			request: apiv1.DeleteSubscriptionRequestObject{Id: publicSubID1},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Delete(gomock.Any(), subID1).Return(nil)
			},
			want: apiv1.DeleteSubscription204Response{},
		},
		{
			name:    "subscription not found",
			request: apiv1.DeleteSubscriptionRequestObject{Id: publicSubID1},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Delete(gomock.Any(), subID1).
					Return(fmt.Errorf("service: %w", domain.ErrSubscriptionNotFound))
			},
			want: apiv1.DeleteSubscription404JSONResponse(errSubscriptionNotFound),
		},
		{
			name:    "internal error",
			request: apiv1.DeleteSubscriptionRequestObject{Id: publicSubID1},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Delete(gomock.Any(), subID1).Return(errService)
			},
			want: apiv1.DeleteSubscription500JSONResponse(common.InternalError()),
		},
		{
			name:    "invalid id",
			request: apiv1.DeleteSubscriptionRequestObject{Id: publicEndpointID},
			want:    apiv1.DeleteSubscription404JSONResponse(errSubscriptionNotFound),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := NewMockInterface(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(service)
			}

			resp, err := NewHandler(service).DeleteSubscription(context.Background(), tt.request)

			require.NoError(t, err)
			require.Equal(t, tt.want, resp)
		})
	}
}
