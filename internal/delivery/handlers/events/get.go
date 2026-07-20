package events

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"time"
	"webhook-delivery/internal/delivery/utils"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
)

type GetResponse struct {
	ID        string
	Type      string
	Payload   json.RawMessage
	CreatedAt time.Time
}

type EventGetter interface {
	Get(ctx context.Context, eventID string) (*dto.GetEventResult, error)
}

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
