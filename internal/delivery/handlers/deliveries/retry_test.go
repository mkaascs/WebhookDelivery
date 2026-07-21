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
)

func Test_Retry(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name       string
		id         string
		setupMock  func(m *MockDeliveryRetryer)
		wantStatus int
	}{
		{
			name: "success",
			id:   "del-1",
			setupMock: func(m *MockDeliveryRetryer) {
				m.EXPECT().Retry(gomock.Any(), "del-1").Return(nil)
			},
			wantStatus: http.StatusAccepted,
		},
		{
			name:       "missing id param",
			id:         "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "delivery not found",
			id:   "del-404",
			setupMock: func(m *MockDeliveryRetryer) {
				m.EXPECT().Retry(gomock.Any(), "del-404").Return(domain.ErrDeliveryNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "generic internal error",
			id:   "del-1",
			setupMock: func(m *MockDeliveryRetryer) {
				m.EXPECT().Retry(gomock.Any(), "del-1").Return(errors.New("db down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			retryer := NewMockDeliveryRetryer(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(retryer)
			}

			req := httptest.NewRequest(http.MethodPost, "/deliveries/retry", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			Retry(retryer, log).ServeHTTP(rr, req)

			require.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}
