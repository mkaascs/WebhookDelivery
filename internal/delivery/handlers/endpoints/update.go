package endpoints

import (
	"context"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"webhook-delivery/internal/delivery/middlewares"
	"webhook-delivery/internal/delivery/utils"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
)

type UpdateRequest struct {
	URL         *string `json:"url"`
	IsActive    *bool   `json:"is_active"`
	Description *string `json:"description"`
}

type EndpointUpdater interface {
	Update(ctx context.Context, command dto.UpdateEndpointCommand) error
}

func Update(updater EndpointUpdater, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		const fn = "handlers.endpoints.Update"
		log = log.With(slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(req.Context())))

		payload, ok := middlewares.GetParsedRequestBody[UpdateRequest](req.Context())
		if !ok {
			log.Info("request body is empty")
			render.Status(req, http.StatusNoContent)
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
			URL:         payload.URL,
			IsActive:    payload.IsActive,
			Description: payload.Description,
		})

		if err != nil {
			const msg = "failed to update endpoint"
			if utils.IsCtxError(err) || utils.TryRenderEndpointsError(w, req, err) {
				log.Info(msg, sloglib.Error(err))
				return
			}

			log.Error(msg, sloglib.Error(err), slog.String("url", *payload.URL))
			utils.RenderError(w, req, http.StatusInternalServerError, "internal server error")
			return
		}

		render.Status(req, http.StatusNoContent)
	}
}
