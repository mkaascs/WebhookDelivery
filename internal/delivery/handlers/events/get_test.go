package events

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
		setupMock  func(m *MockEventGetter)
		wantStatus int
	}{
		{
			name:  "success",
			query: "id=ev-1",
			setupMock: func(m *MockEventGetter) {
				m.EXPECT().Get(gomock.Any(), "ev-1").
					Return(&dto.GetEventResult{Event: domain.Event{ID: "ev-1", Type: "order.created"}}, nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing id param",
			query:      "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "event not found",
			query: "id=ev-404",
			setupMock: func(m *MockEventGetter) {
				m.EXPECT().Get(gomock.Any(), "ev-404").Return(nil, domain.ErrEventNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:  "generic internal error",
			query: "id=ev-1",
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

			req := httptest.NewRequest(http.MethodGet, "/events?"+tt.query, nil)
			rr := httptest.NewRecorder()

			Get(getter, log).ServeHTTP(rr, req)

			require.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}
