package endpoints

import (
	"context"
	"log/slog"
	"net/http"
	"webhook-delivery/internal/delivery/middlewares"
	"webhook-delivery/internal/delivery/utils"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type UpdateRequest struct {
	URL         *string `json:"url" example:"https://example.com/webhooks"` // New receiver URL (SSRF-validated if present)
	IsActive    *bool   `json:"is_active" example:"false"`                  // Pause or resume the endpoint
	Description *string `json:"description" example:"Billing webhook"`      // New description
}

type EndpointUpdater interface {
	Update(ctx context.Context, command dto.UpdateEndpointCommand) error
}

// Update godoc
//
//	@Summary		Update an endpoint
//	@Description	Partially updates the URL, active state or description. An empty body is treated as a no-op. If a URL is provided it is SSRF-validated.
//	@Tags			endpoints
//	@Accept			json
//	@Produce		json
//	@Param			id		path	string			true	"Endpoint ID"
//	@Param			request	body	UpdateRequest	true	"Fields to update (all optional)"
//	@Success		204		"No Content"
//	@Failure		400		{object}	utils.Response	"id is not provided or invalid url"
//	@Failure		404		{object}	utils.Response	"endpoint not found"
//	@Failure		500		{object}	utils.Response
//	@Router			/endpoints/{id} [patch]
func Update(updater EndpointUpdater, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		const fn = "handlers.endpoints.Update"
		log = log.With(slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(req.Context())))

		id := chi.URLParam(req, "id")
		if id == "" {
			log.Info("endpoint id is not provided")
			utils.RenderError(w, req, http.StatusBadRequest, "endpoint id is not provided")
			return
		}

		payload, ok := middlewares.GetParsedRequestBody[UpdateRequest](req.Context())
		if !ok {
			log.Info("request body is empty")
			w.WriteHeader(http.StatusNoContent)
			return
		}

		if payload.URL != nil {
			if err := utils.ValidateURL(*payload.URL); err != nil {
				log.Info("validation error", sloglib.Error(err), slog.String("url", *payload.URL))
				utils.RenderError(w, req, http.StatusBadRequest, "invalid url")
				return
			}
		}

		err := updater.Update(req.Context(), dto.UpdateEndpointCommand{
			ID:          id,
			URL:         payload.URL,
			IsActive:    payload.IsActive,
			Description: payload.Description,
		})

		if err != nil {
			const msg = "failed to update endpoint"
			if utils.IsCtxError(err) || utils.TryRenderDomainError(w, req, err) {
				log.Info(msg, sloglib.Error(err))
				return
			}

			log.Error(msg, sloglib.Error(err))
			utils.RenderError(w, req, http.StatusInternalServerError, "internal server error")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
