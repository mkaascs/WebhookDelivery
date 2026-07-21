package deliveries

import (
	"context"
	"errors"
	"github.com/go-chi/chi"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"webhook-delivery/internal/delivery/middlewares"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
)

func Test_GetFromEvent(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name       string
		query      string
		id         string
		setupMock  func(m *MockAllDeliveriesGetter)
		wantStatus int
	}{
		{
			name:  "success",
			query: "?page=1&limit=10",
			id:    "ev-1",
			setupMock: func(m *MockAllDeliveriesGetter) {
				m.EXPECT().GetFromEvent(gomock.Any(), dto.GetDeliveriesFromEventCommand{
					EventID: "ev-1",
					Page:    1,
					Limit:   10,
				}).Return(&dto.GetDeliveriesFromEventResult{
					Deliveries: []domain.Delivery{{ID: "del-1", Status: domain.StatusPending}},
				}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "success falls back to default pagination on invalid query",
			query: "?page=abc&limit=-010",
			id:    "ev-1",
			setupMock: func(m *MockAllDeliveriesGetter) {
				m.EXPECT().GetFromEvent(gomock.Any(), dto.GetDeliveriesFromEventCommand{
					EventID: "ev-1",
					Page:    1,
					Limit:   10,
				}).Return(&dto.GetDeliveriesFromEventResult{
					Deliveries: []domain.Delivery{{ID: "del-1", Status: domain.StatusPending}},
				}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "generic internal error",
			query: "?page=1&limit=10",
			id:    "ev-1",
			setupMock: func(m *MockAllDeliveriesGetter) {
				m.EXPECT().GetFromEvent(gomock.Any(), dto.GetDeliveriesFromEventCommand{
					EventID: "ev-1",
					Page:    1,
					Limit:   10,
				}).Return(nil, errors.New("db down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			getter := NewMockAllDeliveriesGetter(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(getter)
			}

			handler := middlewares.NewUrlPaginationParser(log)(GetFromEvent(getter, log))

			req := httptest.NewRequest(http.MethodGet, "/deliveries"+tt.query, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			require.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}
