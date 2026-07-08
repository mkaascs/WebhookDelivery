package middlewares

import (
	"github.com/go-chi/chi/middleware"
	"log/slog"
	"net/http"
	"time"
)

func NewLogger(log *slog.Logger) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		log = log.With(slog.String("component", "middleware/logging"))
		log.Info("logger middleware is enabled")

		return func(w http.ResponseWriter, req *http.Request) {
			entry := log.With(slog.String("method", req.Method),
				slog.String("path", req.URL.Path),
				slog.String("remote", req.RemoteAddr),
				slog.String("user_agent", req.UserAgent()),
				slog.String("request_id", middleware.GetReqID(req.Context())))

			currentTime := time.Now()
			wrw := middleware.NewWrapResponseWriter(w, req.ProtoMajor)

			defer func() {
				entry.Info("request completed", slog.Int("status", wrw.Status()),
					slog.Int("bytes", wrw.BytesWritten()),
					slog.Duration("duration", time.Since(currentTime)))
			}()

			next.ServeHTTP(wrw, req)
		}
	}
}
