package events

import (
	"context"
	apiv1 "webhook-delivery/gen/rest/v1"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"

	"github.com/google/uuid"
)

type Interface interface {
	Publish(ctx context.Context, command dto.PublishEventCommand) (*domain.Event, int, error)
	Get(ctx context.Context, eventID uuid.UUID) (*domain.Event, error)
}

type Handler struct {
	service Interface
}

func NewHandler(service Interface) *Handler {
	return &Handler{service: service}
}

func (h *Handler) PublishEvent(ctx context.Context, request apiv1.PublishEventRequestObject) (apiv1.PublishEventResponseObject, error) {
	event, deliveriesCreated, err := h.service.Publish(ctx, dto.PublishEventCommand{
		Type:    request.Body.Type,
		Payload: request.Body.Payload,
	})

	if err != nil {
		return ToPublishErrorResponse(err), nil
	}

	return apiv1.PublishEvent202JSONResponse{
		Id:                domain.ToPublicEventID(event.ID),
		Type:              event.Type,
		DeliveriesCreated: deliveriesCreated,
		CreatedAt:         event.CreatedAt,
	}, nil
}

func (h *Handler) GetEvent(ctx context.Context, request apiv1.GetEventRequestObject) (apiv1.GetEventResponseObject, error) {
	id, err := domain.ParseEventID(request.Id)
	if err != nil {
		return ToGetErrorResponse(err), nil
	}

	event, err := h.service.Get(ctx, id)
	if err != nil {
		return ToGetErrorResponse(err), nil
	}

	return apiv1.GetEvent200JSONResponse{
		Id:        domain.ToPublicEventID(event.ID),
		Type:      event.Type,
		Payload:   event.Payload,
		CreatedAt: event.CreatedAt,
	}, nil
}
