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
	"webhook-delivery/internal/domain/dto"
)

func Test_Get(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name       string
		query      string
		setupMock  func(m *MockEndpointGetter)
		wantStatus int
	}{
		{
			name:  "success",
			query: "id=ep-123",
			setupMock: func(m *MockEndpointGetter) {
				m.EXPECT().GetByID(gomock.Any(), "ep-123").
					Return(&dto.GetEndpointResult{ID: "ep-123", URL: "https://hooks.example.com", IsActive: true}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing id param",
			query:      "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "endpoint not found",
			query: "id=ep-404",
			setupMock: func(m *MockEndpointGetter) {
				m.EXPECT().GetByID(gomock.Any(), "ep-404").
					Return(nil, domain.ErrEndpointNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:  "generic internal error",
			query: "id=ep-123",
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

			req := httptest.NewRequest(http.MethodGet, "/endpoints?"+tt.query, nil)
			rr := httptest.NewRecorder()

			Get(getter, log).ServeHTTP(rr, req)

			require.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}
