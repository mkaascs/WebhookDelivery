package deliveries

import (
	"context"
	apiv1 "webhook-delivery/gen/rest/v1"
	"webhook-delivery/internal/delivery/rest/common"
	"webhook-delivery/internal/domain/dto"
)

type Interface interface {
	GetByID(ctx context.Context, id string) (*dto.GetDeliveryResult, error)
	GetFromEvent(ctx context.Context, command dto.GetDeliveriesFromEventCommand) (*dto.GetDeliveriesFromEventResult, error)
	Retry(ctx context.Context, id string) error
}

type Handler struct {
	service Interface
}

func NewHandler(service Interface) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetDelivery(ctx context.Context, request apiv1.GetDeliveryRequestObject) (apiv1.GetDeliveryResponseObject, error) {
	result, err := h.service.GetByID(ctx, request.Id)
	if err != nil {
		return ToGetErrorResponse(err), nil
	}

	return apiv1.GetDelivery200JSONResponse{Delivery: ToDeliveryInfo(result.Delivery)}, nil
}

func (h *Handler) GetEventDeliveries(ctx context.Context, request apiv1.GetEventDeliveriesRequestObject) (apiv1.GetEventDeliveriesResponseObject, error) {
	page, limit := common.Pagination(request.Params.Page, request.Params.Limit)

	result, err := h.service.GetFromEvent(ctx, dto.GetDeliveriesFromEventCommand{
		EventID: request.Id,
		Page:    page,
		Limit:   limit,
	})

	if err != nil {
		return ToGetEventDeliveriesErrorResponse(err), nil
	}

	return apiv1.GetEventDeliveries200JSONResponse{
		Deliveries: ToDeliveriesInfo(result.Deliveries),
		Total:      result.Total,
	}, nil
}

func (h *Handler) RetryDelivery(ctx context.Context, request apiv1.RetryDeliveryRequestObject) (apiv1.RetryDeliveryResponseObject, error) {
	if err := h.service.Retry(ctx, request.Id); err != nil {
		return ToRetryErrorResponse(err), nil
	}

	return apiv1.RetryDelivery202Response{}, nil
}
