package subscriptions

import (
	"context"
	"log/slog"
	"net/http"
	"time"
	"webhook-delivery/internal/delivery/middlewares"
	"webhook-delivery/internal/delivery/utils"
	"webhook-delivery/internal/domain/dto"
	sloglib "webhook-delivery/internal/lib/logging/slog"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
)

type SubscriptionInfo struct {
	ID         string    `json:"id" example:"sub_1a2b3c"`
	EndpointID string    `json:"endpoint_id" example:"ep_1a2b3c"`
	EventType  string    `json:"event_type" example:"order.created"`
	CreatedAt  time.Time `json:"created_at" example:"2026-07-01T12:00:00Z"`
}

type SubscribeRequest struct {
	EventTypes []string `json:"event_types" validate:"required" example:"order.created"` // Event types to subscribe the endpoint to
}

type SubscribeResponse struct {
	utils.Response
	EndpointID    string             `json:"endpoint_id" example:"ep_1a2b3c"`
	Subscriptions []SubscriptionInfo `json:"subscriptions"`
}

type SubscriptionAdder interface {
	Add(ctx context.Context, command dto.AddSubscriptionCommand) (*dto.AddSubscriptionResult, error)
}

// Subscribe godoc
//
//	@Summary		Subscribe an endpoint to event types
//	@Description	Adds subscriptions for the given event types. Idempotent — event types the endpoint is already subscribed to are ignored.
//	@Tags			subscriptions
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string				true	"Endpoint ID"
//	@Param			request	body		SubscribeRequest	true	"Event types to subscribe to"
//	@Success		200		{object}	SubscribeResponse
//	@Failure		400		{object}	utils.Response	"id or event types not provided"
//	@Failure		404		{object}	utils.Response	"endpoint not found"
//	@Failure		409		{object}	utils.Response	"subscription already exists"
//	@Failure		500		{object}	utils.Response
//	@Router			/endpoints/{id}/subscriptions [post]
func Subscribe(adder SubscriptionAdder, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		const fn = "handlers.subscriptions.Subscribe"
		log = log.With(slog.String("fn", fn),
			slog.String("request_id", middleware.GetReqID(req.Context())))

		endpointID := chi.URLParam(req, "id")
		if endpointID == "" {
			log.Info("endpoint id is not provided in url")
			utils.RenderError(w, req, http.StatusBadRequest, "endpoint id is not provided in url")
			return
		}

		payload, ok := middlewares.GetParsedRequestBody[SubscribeRequest](req.Context())
		if !ok {
			log.Error("failed to parse request body. make sure that BodyParserMiddleware is enabled")
			utils.RenderError(w, req, http.StatusInternalServerError, "internal server error")
			return
		}

		if len(payload.EventTypes) == 0 {
			log.Info("no event types provided")
			utils.RenderError(w, req, http.StatusBadRequest, "no event types provided")
			return
		}

		result, err := adder.Add(req.Context(), dto.AddSubscriptionCommand{
			EndpointID: endpointID,
			EventTypes: payload.EventTypes,
		})

		if err != nil {
			const msg = "failed to add subscription"
			if utils.IsCtxError(err) || utils.TryRenderEndpointsError(w, req, err) {
				log.Info(msg, sloglib.Error(err))
				return
			}

			log.Error(msg, sloglib.Error(err), slog.String("endpoint_id", endpointID))
			utils.RenderError(w, req, http.StatusInternalServerError, msg)
			return
		}

		subInfo := make([]SubscriptionInfo, 0, len(result.Subscriptions))
		for _, sub := range result.Subscriptions {
			subInfo = append(subInfo, SubscriptionInfo{
				ID:         sub.ID,
				EventType:  sub.EventType,
				CreatedAt:  sub.CreatedAt,
				EndpointID: sub.EndpointID,
			})
		}

		render.Status(req, http.StatusOK)
		render.JSON(w, req, SubscribeResponse{
			EndpointID:    endpointID,
			Subscriptions: subInfo,
		})
	}
}
