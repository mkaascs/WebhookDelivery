package subscriptions

import (
	"context"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"log/slog"
	"net/http"
	"webhook-delivery/internal/delivery/utils"
	sloglib "webhook-delivery/internal/lib/logging/slog"
)

type SubscriptionDeleter interface {
	Delete(ctx context.Context, id string) error
}

func Delete(deleter SubscriptionDeleter, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		const fn = "handlers.subscriptions.Delete"
		log = log.With(slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(req.Context())))

		subscriptionID := chi.URLParam(req, "id")
		if subscriptionID == "" {
			log.Info("subscription id is not provided")
			utils.RenderError(w, req, http.StatusBadRequest, "subscription id is not provided")
			return
		}

		if err := deleter.Delete(req.Context(), subscriptionID); err != nil {
			const msg = "failed to delete subscription"
			if utils.IsCtxError(err) || utils.TryRenderEndpointsError(w, req, err) {
				log.Info(msg, sloglib.Error(err))
				return
			}

			log.Error(msg, sloglib.Error(err), slog.String("subscription_id", subscriptionID))
			utils.RenderError(w, req, http.StatusInternalServerError, "internal server error")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
