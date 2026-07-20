package middlewares

import (
	"errors"
	"github.com/go-playground/validator"
	"log/slog"
	"net/http"
	"webhook-delivery/internal/delivery/utils"
	sloglib "webhook-delivery/internal/lib/logging/slog"
)

func NewValidator[T any](log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log = log.With(slog.String("component", "middleware/validating"))

		validate := validator.New()
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			payload, ok := GetParsedRequestBody[T](req.Context())
			if !ok {
				log.Warn("request payload is empty. make sure that BodyParserMiddleware is enabled")
				next.ServeHTTP(w, req)
				return
			}

			if err := validate.Struct(payload); err != nil {
				var validationErrs validator.ValidationErrors
				errors.As(err, &validationErrs)
				log.Info("payload is invalid", sloglib.Error(err))
				utils.RenderValidationErrors(w, req, validationErrs)
				return
			}

			next.ServeHTTP(w, req)
		})
	}
}
