package deliveries

import (
	"context"
	apiv1 "webhook-delivery/gen/rest/v1"
	"webhook-delivery/internal/delivery/rest/common"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"

	"github.com/google/uuid"
)

type Interface interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Delivery, error)
	GetFromEvent(ctx context.Context, command dto.GetDeliveriesFromEventCommand) ([]domain.Delivery, int, error)
	Retry(ctx context.Context, id uuid.UUID) error
}

type Handler struct {
	service Interface
}

func NewHandler(service Interface) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetDelivery(ctx context.Context, request apiv1.GetDeliveryRequestObject) (apiv1.GetDeliveryResponseObject, error) {
	id, err := domain.ParseDeliveryID(request.Id)
	if err != nil {
		return ToGetErrorResponse(err), nil
	}

	delivery, err := h.service.GetByID(ctx, id)
	if err != nil {
		return ToGetErrorResponse(err), nil
	}

	return apiv1.GetDelivery200JSONResponse{Delivery: ToDeliveryInfo(*delivery)}, nil
}

func (h *Handler) GetEventDeliveries(ctx context.Context, request apiv1.GetEventDeliveriesRequestObject) (apiv1.GetEventDeliveriesResponseObject, error) {
	eventID, err := domain.ParseEventID(request.Id)
	if err != nil {
		return ToGetEventDeliveriesErrorResponse(err), nil
	}

	page, limit := common.Pagination(request.Params.Page, request.Params.Limit)

	deliveries, total, err := h.service.GetFromEvent(ctx, dto.GetDeliveriesFromEventCommand{
		EventID: eventID,
		Page:    page,
		Limit:   limit,
	})

	if err != nil {
		return ToGetEventDeliveriesErrorResponse(err), nil
	}

	return apiv1.GetEventDeliveries200JSONResponse{
		Deliveries: ToDeliveriesInfo(deliveries),
		Total:      total,
	}, nil
}

func (h *Handler) RetryDelivery(ctx context.Context, request apiv1.RetryDeliveryRequestObject) (apiv1.RetryDeliveryResponseObject, error) {
	id, err := domain.ParseDeliveryID(request.Id)
	if err != nil {
		return ToRetryErrorResponse(err), nil
	}

	if err := h.service.Retry(ctx, id); err != nil {
		return ToRetryErrorResponse(err), nil
	}

	return apiv1.RetryDelivery202Response{}, nil
}
