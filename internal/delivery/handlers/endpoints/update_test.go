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
	"webhook-delivery/internal/mocks"
)

func Test_Update(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name       string
		body       string
		setupMock  func(m *mocks.MockEndpointUpdater)
		wantStatus int
	}{
		{
			name: "success",
			body: `{"url":"https://hooks.example.com/hook","is_active":false}`,
			setupMock: func(m *mocks.MockEndpointUpdater) {
				m.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "invalid url",
			body:       `{"url":"http://127.0.0.1/hook"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "endpoint not found",
			body: `{"url":"https://hooks.example.com/hook"}`,
			setupMock: func(m *mocks.MockEndpointUpdater) {
				m.EXPECT().Update(gomock.Any(), gomock.Any()).Return(domain.ErrEndpointNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "generic internal error",
			body: `{"url":"https://hooks.example.com/hook"}`,
			setupMock: func(m *mocks.MockEndpointUpdater) {
				m.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.New("redis down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			updater := mocks.NewMockEndpointUpdater(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(updater)
			}

			handler := middlewares.NewBodyParser[UpdateRequest](log)(Update(updater, log))

			req := httptest.NewRequest(http.MethodPatch, "/endpoints", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			require.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}
