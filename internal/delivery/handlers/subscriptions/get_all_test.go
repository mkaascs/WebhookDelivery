package subscriptions

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
)

func Test_GetAll(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	createdAt := time.Date(2026, time.July, 1, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		endpointID string
		setupMock  func(m *MockSubscriptionGetter)
		wantStatus int
		wantSubs   []SubscriptionInfo
	}{
		{
			name:       "success",
			endpointID: "ep-1",
			setupMock: func(m *MockSubscriptionGetter) {
				m.EXPECT().GetAll(gomock.Any(), "ep-1").
					Return([]dto.GetSubscriptionResult{
						{ID: "sub-1", EndpointID: "ep-1", EventType: "order.created", CreatedAt: createdAt},
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
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "endpoint not found",
			endpointID: "ep-404",
			setupMock: func(m *MockSubscriptionGetter) {
				m.EXPECT().GetAll(gomock.Any(), "ep-404").Return(nil, domain.ErrEndpointNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "generic internal error",
			endpointID: "ep-1",
			setupMock: func(m *MockSubscriptionGetter) {
				m.EXPECT().GetAll(gomock.Any(), "ep-1").Return(nil, errors.New("redis down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			getter := NewMockSubscriptionGetter(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(getter)
			}

			req := httptest.NewRequest(http.MethodGet, "/subscriptions", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.endpointID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			GetAll(getter, log).ServeHTTP(rr, req)

			require.Equal(t, tt.wantStatus, rr.Code)

			if tt.wantSubs != nil {
				var resp GetAllResponse
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
				require.Equal(t, tt.wantSubs, resp.Subscriptions)
			}
		})
	}
}
