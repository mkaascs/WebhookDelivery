package middlewares

import (
	"context"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"webhook-delivery/internal/delivery/utils"
	sloglib "webhook-delivery/internal/lib/logging/slog"
)

type parsedBodyCtxKey int

const parsedBodyKey parsedBodyCtxKey = 0

const maxBytes int64 = 2 << 20

func NewBodyParser[T any](log *slog.Logger) func(http.Handler) http.HandlerFunc {
	return func(next http.Handler) http.HandlerFunc {
		log = log.With(slog.String("component", "middleware/parsing"))

		methodsToSkip := map[string]bool{
			http.MethodGet:     true,
			http.MethodHead:    true,
			http.MethodOptions: true,
			http.MethodDelete:  true,
		}

		return func(w http.ResponseWriter, req *http.Request) {
			req.Body = http.MaxBytesReader(w, req.Body, maxBytes)

			if methodsToSkip[req.Method] {
				next.ServeHTTP(w, req)
				return
			}

			if req.ContentLength <= 0 {
				log.Info("request body is empty")
				utils.RenderError(w, req, http.StatusBadRequest, "request body is empty")
				return
			}

			var payload T
			if err := render.DecodeJSON(req.Body, &payload); err != nil {
				log.Info("failed to decode request body", sloglib.Error(err))
				utils.RenderError(w, req, http.StatusBadRequest, "failed to decode request body")
				return
			}

			ctx := context.WithValue(req.Context(), parsedBodyKey, payload)
			next.ServeHTTP(w, req.WithContext(ctx))
		}
	}
}

func GetParsedRequestBody[T any](ctx context.Context) (T, bool) {
	payload, ok := ctx.Value(parsedBodyKey).(T)
	return payload, ok
}
