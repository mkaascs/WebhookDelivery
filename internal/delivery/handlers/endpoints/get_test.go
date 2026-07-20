package endpoints

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
		setupMock  func(m *MockEndpointGetter)
		wantStatus int
	}{
		{
			name: "success",
			id:   "ep-123",
			setupMock: func(m *MockEndpointGetter) {
				m.EXPECT().GetByID(gomock.Any(), "ep-123").
					Return(&dto.GetEndpointResult{ID: "ep-123", URL: "https://hooks.example.com", IsActive: true}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing id param",
			id:         "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "endpoint not found",
			id:   "ep-404",
			setupMock: func(m *MockEndpointGetter) {
				m.EXPECT().GetByID(gomock.Any(), "ep-404").
					Return(nil, domain.ErrEndpointNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "generic internal error",
			id:   "ep-123",
			setupMock: func(m *MockEndpointGetter) {
				m.EXPECT().GetByID(gomock.Any(), "ep-123").
					Return(nil, errors.New("redis down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			getter := NewMockEndpointGetter(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(getter)
			}

			req := httptest.NewRequest(http.MethodGet, "/endpoints", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			Get(getter, log).ServeHTTP(rr, req)

			require.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}
