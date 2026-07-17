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
	"webhook-delivery/internal/delivery/middlewares"
	"webhook-delivery/internal/domain/dto"
)

func Test_GetAll(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name       string
		query      string
		setupMock  func(m *MockAllEndpointsGetter)
		wantStatus int
	}{
		{
			name:  "success with explicit pagination",
			query: "page=2&limit=50",
			setupMock: func(m *MockAllEndpointsGetter) {
				m.EXPECT().
					GetAll(gomock.Any(), dto.GetAllEndpointsCommand{Page: 2, Limit: 50}).
					Return(&dto.GetAllEndpointsResult{
						Total:     1,
						Endpoints: []dto.GetEndpointResult{{ID: "ep-123"}},
					}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "success falls back to default pagination on invalid query",
			query: "page=abc&limit=-5",
			setupMock: func(m *MockAllEndpointsGetter) {
				m.EXPECT().
					GetAll(gomock.Any(), dto.GetAllEndpointsCommand{Page: 1, Limit: 10}).
					Return(&dto.GetAllEndpointsResult{Total: 0, Endpoints: nil}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "generic internal error",
			query: "",
			setupMock: func(m *MockAllEndpointsGetter) {
				m.EXPECT().GetAll(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("redis down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			getter := NewMockAllEndpointsGetter(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(getter)
			}

			handler := middlewares.NewUrlPaginationParser(log)(GetAll(getter, log))

			req := httptest.NewRequest(http.MethodGet, "/endpoints?"+tt.query, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			require.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}
