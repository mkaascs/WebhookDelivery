package endpoints

import (
	"context"
	"log/slog"
	"net/http"
	"webhook-delivery/internal/delivery/utils"
	sloglib "webhook-delivery/internal/lib/logging/slog"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type EndpointDeleter interface {
	Delete(ctx context.Context, id string) error
}

// Delete godoc
//
//	@Summary	Delete an endpoint
//	@Tags		endpoints
//	@Produce	json
//	@Param		id	path	string	true	"Endpoint ID"
//	@Success	204	"No Content"
//	@Failure	400	{object}	utils.Response	"id is not provided"
//	@Failure	404	{object}	utils.Response	"endpoint not found"
//	@Failure	500	{object}	utils.Response
//	@Router		/endpoints/{id} [delete]
func Delete(deleter EndpointDeleter, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		const fn = "handlers.endpoints.Delete"
		log = log.With(slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(req.Context())))

		id := chi.URLParam(req, "id")
		if id == "" {
			log.Info("no endpoint id provided")
			utils.RenderError(w, req, http.StatusBadRequest, "url param endpoint id is not provided")
			return
		}

		if err := deleter.Delete(req.Context(), id); err != nil {
			const msg = "failed to delete endpoint"
			if utils.IsCtxError(err) || utils.TryRenderDomainError(w, req, err) {
				log.Info(msg, sloglib.Error(err))
				return
			}

			log.Error(msg, sloglib.Error(err), slog.String("endpoint_id", id))
			utils.RenderError(w, req, http.StatusInternalServerError, "internal server error")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
