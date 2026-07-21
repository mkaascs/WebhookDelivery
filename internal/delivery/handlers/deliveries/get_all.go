package deliveries

import (
	"context"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"webhook-delivery/internal/delivery/utils"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
)

type GetFromEventResponse struct {
	Deliveries []DeliveryInfo `json:"deliveries"`
}

type AllDeliveriesGetter interface {
	GetFromEvent(ctx context.Context, eventID string) ([]dto.GetDeliveryResult, error)
}

func GetFromEvent(getter AllDeliveriesGetter, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		const fn = "handlers.deliveries.GetFromEvent"
		log = log.With(slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(req.Context())))

		eventID := req.URL.Query().Get("event_id")
		if eventID == "" {
			log.Info("no event id provided")
			utils.RenderError(w, req, http.StatusBadRequest, "url param event id is not provided")
			return
		}

		results, err := getter.GetFromEvent(req.Context(), eventID)
		if err != nil {
			const msg = "failed to get deliveries from event"
			if utils.IsCtxError(err) {
				log.Info(msg, sloglib.Error(err), slog.String("event_id", eventID))
				return
			}

			log.Error(msg, sloglib.Error(err), slog.String("event_id", eventID))
			utils.RenderError(w, req, http.StatusInternalServerError, "internal server error")
			return
		}

		deliveries := make([]DeliveryInfo, 0, len(results))
		for _, result := range results {
			deliveries = append(deliveries, mapResultToResponse(result))
		}

		render.Status(req, http.StatusOK)
		render.JSON(w, req, GetFromEventResponse{
			Deliveries: deliveries,
		})
	}
}
