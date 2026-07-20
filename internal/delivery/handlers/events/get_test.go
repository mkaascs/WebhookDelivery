package events

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
)

func Test_Get(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name       string
		id         string
		setupMock  func(m *MockEventGetter)
		wantStatus int
	}{
		{
			name: "success",
			id:   "ev-1",
			setupMock: func(m *MockEventGetter) {
				m.EXPECT().Get(gomock.Any(), "ev-1").
					Return(&dto.GetEventResult{Event: domain.Event{ID: "ev-1", Type: "order.created"}}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing id param",
			id:         "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "event not found",
			id:   "ev-404",
			setupMock: func(m *MockEventGetter) {
				m.EXPECT().Get(gomock.Any(), "ev-404").Return(nil, domain.ErrEventNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "generic internal error",
			id:   "ev-1",
			setupMock: func(m *MockEventGetter) {
				m.EXPECT().Get(gomock.Any(), "ev-1").Return(nil, errors.New("redis down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			getter := NewMockEventGetter(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(getter)
			}

			req := httptest.NewRequest(http.MethodGet, "/events", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			Get(getter, log).ServeHTTP(rr, req)

			require.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}
