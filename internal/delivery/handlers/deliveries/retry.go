package deliveries

import (
	"context"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"log/slog"
	"net/http"
	"webhook-delivery/internal/delivery/utils"
	sloglib "webhook-delivery/internal/lib/logging/slog"
)

type DeliveryRetryer interface {
	Retry(ctx context.Context, id string) error
}

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
