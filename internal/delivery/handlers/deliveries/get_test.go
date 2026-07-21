package deliveries

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
		setupMock  func(m *MockDeliveryGetter)
		wantStatus int
	}{
		{
			name: "success",
			id:   "del-1",
			setupMock: func(m *MockDeliveryGetter) {
				m.EXPECT().GetByID(gomock.Any(), "del-1").
					Return(&dto.GetDeliveryResult{Delivery: domain.Delivery{ID: "del-1", Status: domain.StatusDelivered}}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing id param",
			id:         "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "delivery not found",
			id:   "del-404",
			setupMock: func(m *MockDeliveryGetter) {
				m.EXPECT().GetByID(gomock.Any(), "del-404").Return(nil, domain.ErrDeliveryNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "generic internal error",
			id:   "del-1",
			setupMock: func(m *MockDeliveryGetter) {
				m.EXPECT().GetByID(gomock.Any(), "del-1").Return(nil, errors.New("db down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			getter := NewMockDeliveryGetter(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(getter)
			}

			req := httptest.NewRequest(http.MethodGet, "/deliveries", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			Get(getter, log).ServeHTTP(rr, req)

			require.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}
