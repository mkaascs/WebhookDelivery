package subscriptions

import (
	"context"
	"log/slog"
	"net/http"
	"webhook-delivery/internal/delivery/utils"
	sloglib "webhook-delivery/internal/lib/logging/slog"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type SubscriptionDeleter interface {
	Delete(ctx context.Context, id string) error
}

// Delete godoc
//
//	@Summary	Delete a subscription
//	@Tags		subscriptions
//	@Produce	json
//	@Param		id	path	string	true	"Subscription ID"
//	@Success	204	"No Content"
//	@Failure	400	{object}	utils.Response	"id is not provided"
//	@Failure	404	{object}	utils.Response	"subscription not found"
//	@Failure	500	{object}	utils.Response
//	@Router		/subscriptions/{id} [delete]
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
			if utils.IsCtxError(err) || utils.TryRenderDomainError(w, req, err) {
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
