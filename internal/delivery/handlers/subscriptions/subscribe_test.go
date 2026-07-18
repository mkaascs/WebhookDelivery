package subscriptions

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"webhook-delivery/internal/delivery/middlewares"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
)

func Test_Subscribe(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	createdAt := time.Date(2026, time.July, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		endpointID string
		body       string
		setupMock  func(m *MockSubscriptionAdder)
		wantStatus int
		wantSubs   []SubscriptionInfo
	}{
		{
			name:       "success",
			endpointID: "ep-1",
			body:       `{"event_types":["order.created"]}`,
			setupMock: func(m *MockSubscriptionAdder) {
				m.EXPECT().
					Add(gomock.Any(), dto.AddSubscriptionCommand{EndpointID: "ep-1", EventTypes: []string{"order.created"}}).
					Return(&dto.AddSubscriptionResult{
						Subscriptions: []domain.Subscription{
							{ID: "sub-1", EndpointID: "ep-1", EventType: "order.created", CreatedAt: createdAt},
						},
					}, nil)
			},
			wantStatus: http.StatusOK,
			wantSubs: []SubscriptionInfo{
				{ID: "sub-1", EndpointID: "ep-1", EventType: "order.created", CreatedAt: createdAt},
			},
		},
		{
			name:       "missing endpoint id",
			endpointID: "",
			body:       `{"event_types":["order.created"]}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "no event types",
			endpointID: "ep-1",
			body:       `{"event_types":[]}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "endpoint not found",
			endpointID: "ep-404",
			body:       `{"event_types":["order.created"]}`,
			setupMock: func(m *MockSubscriptionAdder) {
				m.EXPECT().Add(gomock.Any(), gomock.Any()).Return(nil, domain.ErrEndpointNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "subscription already exists",
			endpointID: "ep-1",
			body:       `{"event_types":["order.created"]}`,
			setupMock: func(m *MockSubscriptionAdder) {
				m.EXPECT().Add(gomock.Any(), gomock.Any()).Return(nil, domain.ErrSubscriptionAlreadyExists)
			},
			wantStatus: http.StatusConflict,
		},
		{
			name:       "generic internal error",
			endpointID: "ep-1",
			body:       `{"event_types":["order.created"]}`,
			setupMock: func(m *MockSubscriptionAdder) {
				m.EXPECT().Add(gomock.Any(), gomock.Any()).Return(nil, errors.New("redis down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			adder := NewMockSubscriptionAdder(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(adder)
			}

			handler := middlewares.NewBodyParser[SubscribeRequest](log)(Subscribe(adder, log))

			req := httptest.NewRequest(http.MethodPost, "/endpoints/subscriptions", strings.NewReader(tt.body))
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.endpointID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tt.wantStatus, rr.Code)

			if tt.wantSubs != nil {
				var resp SubscribeResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
				require.Equal(t, tt.endpointID, resp.EndpointID)
				require.Equal(t, tt.wantSubs, resp.Subscriptions)
			}
		})
	}
}
