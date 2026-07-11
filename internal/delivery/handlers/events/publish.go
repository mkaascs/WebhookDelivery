package events

import (
	"context"
	"encoding/json"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"log/slog"
	"net/http"
	"time"
	"webhook-delivery/internal/delivery/middlewares"
	"webhook-delivery/internal/delivery/utils"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"
)

type PublishRequest struct {
	Type    string          `json:"type" validate:"required"`
	Payload json.RawMessage `json:"payload"`
}

type PublishResponse struct {
	utils.Response
	ID                string    `json:"id"`
	Type              string    `json:"type"`
	CreatedAt         time.Time `json:"created_at"`
	DeliveriesCreated int       `json:"deliveries_created"`
}

type EventPublisher interface {
	Publish(ctx context.Context, command dto.PublishEventCommand) (*dto.PublishEventResult, error)
}

func Publish(publisher EventPublisher, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		const fn = "handlers.events.Publish"
		log = log.With(slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(req.Context())))

		payload, ok := middlewares.GetParsedRequestBody[PublishRequest](req.Context())
		if !ok {
			log.Error("failed to parse request body. make sure that BodyParserMiddleware is enabled")
			utils.RenderError(w, req, http.StatusInternalServerError, "internal server error")
			return
		}

		result, err := publisher.Publish(req.Context(), dto.PublishEventCommand{
			Type:    payload.Type,
			Payload: payload.Payload,
		})

		if err != nil {
			const msg = "failed to publish event"
			if utils.IsCtxError(err) {
				log.Info(msg, sloglib.Error(err))
				return
			}

			log.Error(msg, sloglib.Error(err), slog.String("event_type", payload.Type))
			utils.RenderError(w, req, http.StatusInternalServerError, "internal server error")
			return
		}

		render.Status(req, http.StatusAccepted)
		render.JSON(w, req, PublishResponse{
			ID:                result.Event.ID,
			Type:              result.Event.Type,
			CreatedAt:         result.Event.CreatedAt,
			DeliveriesCreated: result.DeliveriesCreated,
		})
	}
}
