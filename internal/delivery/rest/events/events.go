package events

import (
	"context"
	apiv1 "webhook-delivery/gen/rest/v1"
	"webhook-delivery/internal/domain/dto"
)

type Interface interface {
	Publish(ctx context.Context, command dto.PublishEventCommand) (*dto.PublishEventResult, error)
	Get(ctx context.Context, eventID string) (*dto.GetEventResult, error)
}

type Handler struct {
	service Interface
}

func NewHandler(service Interface) *Handler {
	return &Handler{service: service}
}

func (h *Handler) PublishEvent(ctx context.Context, request apiv1.PublishEventRequestObject) (apiv1.PublishEventResponseObject, error) {
	result, err := h.service.Publish(ctx, dto.PublishEventCommand{
		Type:    request.Body.Type,
		Payload: request.Body.Payload,
	})

	if err != nil {
		return ToPublishErrorResponse(err), nil
	}

	return apiv1.PublishEvent202JSONResponse{
		Id:                result.Event.ID,
		Type:              result.Event.Type,
		DeliveriesCreated: result.DeliveriesCreated,
		CreatedAt:         result.Event.CreatedAt,
	}, nil
}

func (h *Handler) GetEvent(ctx context.Context, request apiv1.GetEventRequestObject) (apiv1.GetEventResponseObject, error) {
	result, err := h.service.Get(ctx, request.Id)
	if err != nil {
		return ToGetErrorResponse(err), nil
	}

	return apiv1.GetEvent200JSONResponse{
		Id:        result.Event.ID,
		Type:      result.Event.Type,
		Payload:   result.Event.Payload,
		CreatedAt: result.Event.CreatedAt,
	}, nil
}
