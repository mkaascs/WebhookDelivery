package endpoints

import (
	"context"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"time"
	"webhook-delivery/internal/delivery/middlewares"
	"webhook-delivery/internal/delivery/utils"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
)

type RegisterRequest struct {
	URL         string   `json:"url" validate:"url,required"`
	EventTypes  []string `json:"event_types"`
	Description *string  `json:"description"`
}

type RegisterResponse struct {
	utils.Response
	ID         string    `json:"id"`
	URL        string    `json:"url"`
	EventTypes []string  `json:"event_types,omitempty"`
	Secret     string    `json:"secret"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

type EndpointRegistrar interface {
	Register(ctx context.Context, command dto.RegisterEndpointCommand) (*dto.RegisterEndpointResult, error)
}

func Register(registrar EndpointRegistrar, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		const fn = "handlers.endpoints.Register"
		log = log.With(slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(req.Context())))

		payload, ok := middlewares.GetParsedRequestBody[RegisterRequest](req.Context())
		if !ok {
			log.Error("failed to parse request body. make sure that BodyParserMiddleware is enabled")
			utils.RenderError(w, req, http.StatusInternalServerError, "internal server error")
			return
		}

		result, err := registrar.Register(req.Context(), dto.RegisterEndpointCommand{
			URL:         payload.URL,
			EventTypes:  payload.EventTypes,
			Description: payload.Description,
		})

		if err != nil {
			log.Error("failed to register endpoint", sloglib.Error(err), slog.String("url", payload.URL))
			utils.RenderError(w, req, http.StatusInternalServerError, "internal server error")
			return
		}

		render.Status(req, http.StatusOK)
		render.JSON(w, req, RegisterResponse{
			ID:         result.ID,
			URL:        payload.URL,
			EventTypes: payload.EventTypes,
			Secret:     result.Secret,
			IsActive:   result.IsActive,
			CreatedAt:  result.CreatedAt,
		})
	}
}
