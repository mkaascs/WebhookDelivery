package subscriptions

import (
	"context"
	"log/slog"
	"net/http"
	"webhook-delivery/internal/delivery/utils"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type GetAllResponse struct {
	utils.Response
	Subscriptions []SubscriptionInfo `json:"subscriptions"`
}

type SubscriptionGetter interface {
	GetAll(ctx context.Context, endpointID string) ([]dto.GetSubscriptionResult, error)
}

// GetAll godoc
//
//	@Summary	List an endpoint's subscriptions
//	@Tags		subscriptions
//	@Produce	json
//	@Param		id	path		string	true	"Endpoint ID"
//	@Success	200	{object}	GetAllResponse
//	@Failure	400	{object}	utils.Response	"id is not provided"
//	@Failure	404	{object}	utils.Response	"endpoint not found"
//	@Failure	500	{object}	utils.Response
//	@Router		/endpoints/{id}/subscriptions [get]
func GetAll(getter SubscriptionGetter, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		const fn = "handlers.subscriptions.GetAll"
		log = log.With(slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(req.Context())))

		endpointID := chi.URLParam(req, "id")
		if endpointID == "" {
			log.Info("endpoint id is not provided")
			utils.RenderError(w, req, http.StatusBadRequest, "endpoint id is not provided")
			return
		}

		subs, err := getter.GetAll(req.Context(), endpointID)
		if err != nil {
			const msg = "failed to get all subscriptions"
			if utils.IsCtxError(err) || utils.TryRenderDomainError(w, req, err) {
				log.Info(msg, sloglib.Error(err), slog.String("endpoint_id", endpointID))
				return
			}

			log.Error(msg, sloglib.Error(err), slog.String("endpoint_id", endpointID))
			utils.RenderError(w, req, http.StatusInternalServerError, "internal server error")
			return
		}

		subInfo := make([]SubscriptionInfo, 0, len(subs))
		for _, sub := range subs {
			subInfo = append(subInfo, SubscriptionInfo{
				ID:         sub.ID,
				EndpointID: sub.EndpointID,
				EventType:  sub.EventType,
				CreatedAt:  sub.CreatedAt,
			})
		}

		render.Status(req, http.StatusOK)
		render.JSON(w, req, &GetAllResponse{
			Subscriptions: subInfo,
		})
	}
}
