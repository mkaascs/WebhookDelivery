package endpoints

import (
	"context"
	"log/slog"
	"net/http"
	"time"
	"webhook-delivery/internal/delivery/middlewares"
	"webhook-delivery/internal/delivery/utils"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type RegisterRequest struct {
	URL         string   `json:"url" validate:"url,required" example:"https://example.com/webhooks"` // Receiver URL (http/https, must not be private)
	EventTypes  []string `json:"event_types" example:"order.created"`                                // Event types to subscribe to on creation
	Description *string  `json:"description" example:"Billing webhook"`                              // Optional human-readable description
}

type RegisterResponse struct {
	utils.Response
	ID         string    `json:"id" example:"ep_1a2b3c"`                     // Generated endpoint ID
	URL        string    `json:"url" example:"https://example.com/webhooks"` // Receiver URL
	EventTypes []string  `json:"event_types" example:"order.created"`        // Subscribed event types
	Secret     string    `json:"secret" example:"whsec_3f9a1c2e8d"`          // Signing secret (returned once, on creation)
	IsActive   bool      `json:"is_active" example:"true"`                   // Whether the endpoint is active
	CreatedAt  time.Time `json:"created_at" example:"2026-07-01T12:00:00Z"`  // Creation timestamp
}

type EndpointRegistrar interface {
	Register(ctx context.Context, command dto.RegisterEndpointCommand) (*dto.RegisterEndpointResult, error)
}

// Register godoc
//
//	@Summary		Register a webhook endpoint
//	@Description	Registers a receiver URL and returns its generated signing secret. The URL is SSRF-validated (loopback / link-local / localhost are rejected).
//	@Tags			endpoints
//	@Accept			json
//	@Produce		json
//	@Param			request	body		RegisterRequest	true	"Endpoint to register"
//	@Success		200		{object}	RegisterResponse
//	@Failure		400		{object}	utils.Response	"invalid url or malformed body"
//	@Failure		500		{object}	utils.Response
//	@Router			/endpoints [post]
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

		if err := utils.ValidateURL(payload.URL); err != nil {
			log.Info("validation error", sloglib.Error(err), slog.String("url", payload.URL))
			utils.RenderError(w, req, http.StatusBadRequest, "invalid url")
			return
		}

		result, err := registrar.Register(req.Context(), dto.RegisterEndpointCommand{
			URL:         payload.URL,
			EventTypes:  payload.EventTypes,
			Description: payload.Description,
		})

		if err != nil {
			const msg = "failed to register endpoint"
			if utils.IsCtxError(err) || utils.TryRenderDomainError(w, req, err) {
				log.Info(msg, sloglib.Error(err))
				return
			}

			log.Error(msg, sloglib.Error(err), slog.String("url", payload.URL))
			utils.RenderError(w, req, http.StatusInternalServerError, "internal server error")
			return
		}

		render.Status(req, http.StatusOK)
		render.JSON(w, req, RegisterResponse{
			ID:         result.ID,
			URL:        payload.URL,
			EventTypes: utils.GetDefaultIfNull(payload.EventTypes),
			Secret:     result.Secret,
			IsActive:   result.IsActive,
			CreatedAt:  result.CreatedAt,
		})
	}
}
