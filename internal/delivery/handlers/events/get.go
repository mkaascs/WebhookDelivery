package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
	"webhook-delivery/internal/delivery/utils"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type GetResponse struct {
	ID        string          `json:"id" example:"ev_1a2b3c"`
	Type      string          `json:"type" example:"order.created"`
	Payload   json.RawMessage `json:"payload" swaggertype:"object"`
	CreatedAt time.Time       `json:"created_at" example:"2026-07-01T12:00:00Z"`
}

type EventGetter interface {
	Get(ctx context.Context, eventID string) (*dto.GetEventResult, error)
}

// Get godoc
//
//	@Summary	Get an event by ID
//	@Tags		events
//	@Produce	json
//	@Param		id	path		string	true	"Event ID"
//	@Success	200	{object}	GetResponse
//	@Failure	400	{object}	utils.Response	"id is not provided"
//	@Failure	404	{object}	utils.Response	"event not found"
//	@Failure	500	{object}	utils.Response
//	@Router		/events/{id} [get]
func Get(getter EventGetter, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		const fn = "handlers.events.Get"
		log = log.With(slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(req.Context())))

		id := chi.URLParam(req, "id")
		if id == "" {
			log.Info("event id is not provided")
			utils.RenderError(w, req, http.StatusBadRequest, "event id is not provided")
			return
		}

		result, err := getter.Get(req.Context(), id)
		if err != nil {
			const msg = "failed to get event"
			if utils.IsCtxError(err) || utils.TryRenderEndpointsError(w, req, err) {
				log.Info(msg, sloglib.Error(err))
				return
			}

			log.Error(msg, sloglib.Error(err), slog.String("event_id", id))
			utils.RenderError(w, req, http.StatusInternalServerError, "internal server error")
			return
		}

		render.Status(req, http.StatusOK)
		render.JSON(w, req, GetResponse{
			ID:        result.Event.ID,
			Type:      result.Event.Type,
			Payload:   result.Event.Payload,
			CreatedAt: result.Event.CreatedAt,
		})
	}
}
