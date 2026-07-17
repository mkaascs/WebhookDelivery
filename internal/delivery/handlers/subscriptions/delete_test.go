package subscriptions

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
		setupMock  func(m *MockSubscriptionDeleter)
		wantStatus int
	}{
		{
			name:  "success",
			query: "id=sub-1",
			setupMock: func(m *MockSubscriptionDeleter) {
				m.EXPECT().Delete(gomock.Any(), "sub-1").Return(nil)
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "missing id param",
			query:      "",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "subscription not found",
			query: "id=sub-404",
			setupMock: func(m *MockSubscriptionDeleter) {
				m.EXPECT().Delete(gomock.Any(), "sub-404").Return(domain.ErrSubscriptionNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:  "generic internal error",
			query: "id=sub-1",
			setupMock: func(m *MockSubscriptionDeleter) {
				m.EXPECT().Delete(gomock.Any(), "sub-1").Return(errors.New("redis down"))
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			deleter := NewMockSubscriptionDeleter(ctrl)
			if tt.setupMock != nil {
				tt.setupMock(deleter)
			}

			req := httptest.NewRequest(http.MethodDelete, "/subscriptions?"+tt.query, nil)
			rr := httptest.NewRecorder()

			Delete(deleter, log).ServeHTTP(rr, req)

			require.Equal(t, tt.wantStatus, rr.Code)
		})
	}
}
