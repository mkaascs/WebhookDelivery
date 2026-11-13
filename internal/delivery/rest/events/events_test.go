package events

import (
	"context"
	"encoding/json"
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
	payload    = json.RawMessage(`{"order_id":42}`)
)

func Test_PublishEvent(t *testing.T) {
	tests := []struct {
		name      string
		request   apiv1.PublishEventRequestObject
		setupMock func(m *MockInterface)
		want      apiv1.PublishEventResponseObject
	}{
		{
			name: "success",
			request: apiv1.PublishEventRequestObject{Body: &apiv1.PublishEventRequest{
				Type:    "order.created",
				Payload: payload,
			}},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Publish(gomock.Any(), dto.PublishEventCommand{
					Type:    "order.created",
					Payload: payload,
				}).Return(&dto.PublishEventResult{
					Event:             domain.Event{ID: "ev-1", Type: "order.created", Payload: payload, CreatedAt: now},
					DeliveriesCreated: 3,
				}, nil)
			},
			want: apiv1.PublishEvent202JSONResponse{
				Id:                "ev-1",
				Type:              "order.created",
				DeliveriesCreated: 3,
				CreatedAt:         now,
			},
		},
		{
			name:    "without payload and subscribers",
			request: apiv1.PublishEventRequestObject{Body: &apiv1.PublishEventRequest{Type: "order.created"}},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Publish(gomock.Any(), dto.PublishEventCommand{Type: "order.created"}).
					Return(&dto.PublishEventResult{
						Event: domain.Event{ID: "ev-1", Type: "order.created", CreatedAt: now},
					}, nil)
			},
			want: apiv1.PublishEvent202JSONResponse{
				Id:        "ev-1",
				Type:      "order.created",
				CreatedAt: now,
			},
		},
		{
			name:    "internal error",
			request: apiv1.PublishEventRequestObject{Body: &apiv1.PublishEventRequest{Type: "order.created"}},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Publish(gomock.Any(), gomock.Any()).Return(nil, errService)
			},
			want: apiv1.PublishEvent500JSONResponse(common.InternalError()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := NewMockInterface(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(service)
			}

			resp, err := NewHandler(service).PublishEvent(context.Background(), tt.request)

			require.NoError(t, err)
			require.Equal(t, tt.want, resp)
		})
	}
}

func Test_GetEvent(t *testing.T) {
	tests := []struct {
		name      string
		request   apiv1.GetEventRequestObject
		setupMock func(m *MockInterface)
		want      apiv1.GetEventResponseObject
	}{
		{
			name:    "success",
			request: apiv1.GetEventRequestObject{Id: "ev-1"},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Get(gomock.Any(), "ev-1").Return(&dto.GetEventResult{
					Event: domain.Event{ID: "ev-1", Type: "order.created", Payload: payload, CreatedAt: now},
				}, nil)
			},
			want: apiv1.GetEvent200JSONResponse{
				Id:        "ev-1",
				Type:      "order.created",
				Payload:   payload,
				CreatedAt: now,
			},
		},
		{
			name:    "event not found",
			request: apiv1.GetEventRequestObject{Id: "ev-404"},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Get(gomock.Any(), "ev-404").
					Return(nil, fmt.Errorf("service: %w", domain.ErrEventNotFound))
			},
			want: apiv1.GetEvent404JSONResponse(errEventNotFound),
		},
		{
			name:    "internal error",
			request: apiv1.GetEventRequestObject{Id: "ev-1"},
			setupMock: func(m *MockInterface) {
				m.EXPECT().Get(gomock.Any(), "ev-1").Return(nil, errService)
			},
			want: apiv1.GetEvent500JSONResponse(common.InternalError()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			service := NewMockInterface(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(service)
			}

			resp, err := NewHandler(service).GetEvent(context.Background(), tt.request)

			require.NoError(t, err)
			require.Equal(t, tt.want, resp)
		})
	}
}
