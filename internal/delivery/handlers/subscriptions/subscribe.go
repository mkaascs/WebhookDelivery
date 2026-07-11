package subscriptions

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

type SubscriptionInfo struct {
	ID        string `json:"id"`
	EventType string `json:"event_type"`
}

type SubscribeRequest struct {
	EventTypes []string `json:"event_types" validate:"required"`
}

type SubscribeResponse struct {
	utils.Response
	EndpointID    string             `json:"endpoint_id"`
	Subscriptions []SubscriptionInfo `json:"subscriptions"`
}

type SubscriptionAdder interface {
	Add(ctx context.Context, command dto.AddSubscriptionCommand) (*dto.AddSubscriptionResult, error)
}

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
				ID:        sub.ID,
				EventType: sub.EventType,
			})
		}

		render.Status(req, http.StatusOK)
		render.JSON(w, req, SubscribeResponse{
			EndpointID:    endpointID,
			Subscriptions: subInfo,
		})
	}
}
