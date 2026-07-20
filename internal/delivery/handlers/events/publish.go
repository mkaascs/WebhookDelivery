package events

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
	"webhook-delivery/internal/delivery/middlewares"
	"webhook-delivery/internal/delivery/utils"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type PublishRequest struct {
	Type    string          `json:"type" validate:"required" example:"order.created"` // Event type
	Payload json.RawMessage `json:"payload" swaggertype:"object"`                     // Arbitrary JSON payload delivered to subscribers
}

type PublishResponse struct {
	utils.Response
	ID                string    `json:"id" example:"ev_1a2b3c"`
	Type              string    `json:"type" example:"order.created"`
	CreatedAt         time.Time `json:"created_at" example:"2026-07-01T12:00:00Z"`
	DeliveriesCreated int       `json:"deliveries_created" example:"3"` // Number of pending deliveries created for subscribers
}

type EventPublisher interface {
	Publish(ctx context.Context, command dto.PublishEventCommand) (*dto.PublishEventResult, error)
}

// Publish godoc
//
//	@Summary		Publish an event
//	@Description	Publishes an event and fans it out to subscribed endpoints, creating a pending delivery for each. Returns immediately with the number of deliveries created.
//	@Tags			events
//	@Accept			json
//	@Produce		json
//	@Param			request	body		PublishRequest	true	"Event to publish"
//	@Success		202		{object}	PublishResponse
//	@Failure		400		{object}	utils.Response	"event type is empty or malformed body"
//	@Failure		500		{object}	utils.Response
//	@Router			/events [post]
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

		if payload.Type == "" {
			log.Info("event type is empty")
			utils.RenderError(w, req, http.StatusBadRequest, "event type is empty")
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
