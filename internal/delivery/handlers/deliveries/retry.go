package deliveries

import (
	"context"
	"log/slog"
	"net/http"
	"webhook-delivery/internal/delivery/utils"
	sloglib "webhook-delivery/internal/lib/logging/slog"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type DeliveryRetryer interface {
	Retry(ctx context.Context, id string) error
}

// Retry godoc
//
//	@Summary		Retry a delivery
//	@Description	Resets a delivery to pending and re-queues it for immediate delivery, even if the endpoint is currently inactive.
//	@Tags			deliveries
//	@Produce		json
//	@Param			id	path	string	true	"Delivery ID"
//	@Success		202	"Accepted"
//	@Failure		400	{object}	utils.Response	"id is not provided"
//	@Failure		404	{object}	utils.Response	"delivery not found"
//	@Failure		500	{object}	utils.Response
//	@Router			/deliveries/{id}/retry [post]
func Retry(retryer DeliveryRetryer, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		const fn = "handlers.deliveries.Retry"
		log = log.With(slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(req.Context())))

		id := chi.URLParam(req, "id")
		if id == "" {
			log.Info("no delivery id provided")
			utils.RenderError(w, req, http.StatusBadRequest, "url param delivery id is not provided")
			return
		}

		err := retryer.Retry(req.Context(), id)
		if err != nil {
			const msg = "failed to retry delivery"
			if utils.IsCtxError(err) || utils.TryRenderDomainError(w, req, err) {
				log.Info(msg, sloglib.Error(err), slog.String("id", id))
				return
			}

			log.Error(msg, sloglib.Error(err), slog.String("id", id))
			utils.RenderError(w, req, http.StatusInternalServerError, "internal server error")
			return
		}

		w.WriteHeader(http.StatusAccepted)
	}
}
