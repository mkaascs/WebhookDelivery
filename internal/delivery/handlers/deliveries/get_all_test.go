package deliveries

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

func Test_GetFromEvent(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name       string
		query      string
		setupMock  func(m *MockAllDeliveriesGetter)
		wantStatus int
	}{
		{
			name:  "success",
			query: "event_id=ev-1",
			setupMock: func(m *MockAllDeliveriesGetter) {
				m.EXPECT().GetFromEvent(gomock.Any(), "ev-1").
					Return([]dto.GetDeliveryResult{{Delivery: domain.Delivery{ID: "del-1", Status: domain.StatusPending}}}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing event id param",
			query:      "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "generic internal error",
			query: "event_id=ev-1",
			setupMock: func(m *MockAllDeliveriesGetter) {
				m.EXPECT().GetFromEvent(gomock.Any(), "ev-1").Return(nil, errors.New("db down"))
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

			req := httptest.NewRequest(http.MethodGet, "/deliveries?"+tt.query, nil)
			rr := httptest.NewRecorder()

			GetFromEvent(getter, log).ServeHTTP(rr, req)

			require.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}
