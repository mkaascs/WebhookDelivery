package endpoints

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"webhook-delivery/internal/domain"
)

func Test_Delete(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name       string
		query      string
		setupMock  func(m *MockEndpointDeleter)
		wantStatus int
	}{
		{
			name:  "success",
			query: "id=ep-123",
			setupMock: func(m *MockEndpointDeleter) {
				m.EXPECT().Delete(gomock.Any(), "ep-123").Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "missing id param",
			query:      "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "endpoint not found",
			query: "id=ep-404",
			setupMock: func(m *MockEndpointDeleter) {
				m.EXPECT().Delete(gomock.Any(), "ep-404").Return(domain.ErrEndpointNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:  "generic internal error",
			query: "id=ep-123",
			setupMock: func(m *MockEndpointDeleter) {
				m.EXPECT().Delete(gomock.Any(), "ep-123").Return(errors.New("redis down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			deleter := NewMockEndpointDeleter(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(deleter)
			}

			req := httptest.NewRequest(http.MethodDelete, "/endpoints?"+tt.query, nil)
			rr := httptest.NewRecorder()

			Delete(deleter, log).ServeHTTP(rr, req)

			require.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}
