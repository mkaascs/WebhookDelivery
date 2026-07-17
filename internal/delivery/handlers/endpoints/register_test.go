package endpoints

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"webhook-delivery/internal/delivery/middlewares"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
)

func Test_Register(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name       string
		body       string
		setupMock  func(m *MockEndpointRegistrar)
		wantStatus int
	}{
		{
			name: "success",
			body: `{"url":"https://hooks.example.com/billing","event_types":["order.created"]}`,
			setupMock: func(m *MockEndpointRegistrar) {
				m.EXPECT().Register(gomock.Any(), gomock.Any()).
					Return(&dto.RegisterEndpointResult{ID: "ep-123", Secret: "whsec_x", IsActive: true}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "malformed json",
			body:       `{"url":`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid url",
			body:       `{"url":"http://127.0.0.1/hook"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "domain not found error",
			body: `{"url":"https://hooks.example.com/hook"}`,
			setupMock: func(m *MockEndpointRegistrar) {
				m.EXPECT().Register(gomock.Any(), gomock.Any()).
					Return(nil, domain.ErrEndpointNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "generic internal error",
			body: `{"url":"https://hooks.example.com/hook"}`,
			setupMock: func(m *MockEndpointRegistrar) {
				m.EXPECT().Register(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("redis down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			registrar := NewMockEndpointRegistrar(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(registrar)
			}

			handler := middlewares.NewBodyParser[RegisterRequest](log)(Register(registrar, log))

			req := httptest.NewRequest(http.MethodPost, "/endpoints", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			require.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}
