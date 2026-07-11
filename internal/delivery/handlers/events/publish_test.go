package events

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
	"webhook-delivery/internal/mocks"
)

func Test_Publish(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	tests := []struct {
		name       string
		body       string
		setupMock  func(m *mocks.MockEventPublisher)
		wantStatus int
	}{
		{
			name: "success",
			body: `{"type":"order.created","payload":{"amount":100}}`,
			setupMock: func(m *mocks.MockEventPublisher) {
				m.EXPECT().Publish(gomock.Any(), gomock.Any()).
					Return(&dto.PublishEventResult{
						Event:             domain.Event{ID: "ev-1", Type: "order.created"},
						DeliveriesCreated: 3,
					}, nil)
			},
			wantStatus: http.StatusAccepted,
		},
		{
			name:       "malformed json",
			body:       `{"type":`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "empty event type",
			body:       `{"payload":{"amount":100}}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "generic internal error",
			body: `{"type":"order.created","payload":{"amount":100}}`,
			setupMock: func(m *mocks.MockEventPublisher) {
				m.EXPECT().Publish(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("redis down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			publisher := mocks.NewMockEventPublisher(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(publisher)
			}

			handler := middlewares.NewBodyParser[PublishRequest](log)(Publish(publisher, log))

			req := httptest.NewRequest(http.MethodPost, "/events", strings.NewReader(tt.body))
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			require.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}
