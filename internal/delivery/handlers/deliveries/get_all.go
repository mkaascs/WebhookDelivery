package deliveries

import (
	"context"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"webhook-delivery/internal/delivery/middlewares"
	"webhook-delivery/internal/delivery/utils"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
)

type GetFromEventResponse struct {
	Total      int            `json:"total"`
	Deliveries []DeliveryInfo `json:"deliveries"`
}

type AllDeliveriesGetter interface {
	GetFromEvent(ctx context.Context, command dto.GetDeliveriesFromEventCommand) (*dto.GetDeliveriesFromEventResult, error)
}

func GetFromEvent(getter AllDeliveriesGetter, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		const fn = "handlers.deliveries.GetFromEvent"
		log = log.With(slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(req.Context())))

		eventID := chi.URLParam(req, "id")
		if eventID == "" {
			log.Info("no event id provided")
			utils.RenderError(w, req, http.StatusBadRequest, "url param event id is not provided")
			return
		}

		pagination, ok := middlewares.GetPaginationParams(req.Context())
		if !ok {
			log.Error("failed to get pagination params. make sure that PaginationParserMiddleware is enabled")
			utils.RenderError(w, req, http.StatusInternalServerError, "internal server error")
			return
		}

		results, err := getter.GetFromEvent(req.Context(), dto.GetDeliveriesFromEventCommand{
			EventID: eventID,
			Limit:   pagination.Limit,
			Page:    pagination.Page,
		})

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

		deliveries := make([]DeliveryInfo, 0, len(results.Deliveries))
		for _, result := range results.Deliveries {
			deliveries = append(deliveries, mapDomainToResponse(result))
		}

		render.Status(req, http.StatusOK)
		render.JSON(w, req, GetFromEventResponse{
			Deliveries: deliveries,
			Total:      results.Total,
		})
	}
}
