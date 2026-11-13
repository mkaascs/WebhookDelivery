package subscriptions

import (
	"context"
	apiv1 "webhook-delivery/gen/rest/v1"
	"webhook-delivery/internal/domain/dto"
)

type Interface interface {
	GetAll(ctx context.Context, endpointID string) ([]dto.GetSubscriptionResult, error)
	Add(ctx context.Context, command dto.AddSubscriptionCommand) (*dto.AddSubscriptionResult, error)
	Delete(ctx context.Context, id string) error
}

type Handler struct {
	service Interface
}

func NewHandler(service Interface) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetAllSubscriptions(ctx context.Context, request apiv1.GetAllSubscriptionsRequestObject) (apiv1.GetAllSubscriptionsResponseObject, error) {
	result, err := h.service.GetAll(ctx, request.Id)
	if err != nil {
		return ToGetAllErrorResponse(err), nil
	}

	return apiv1.GetAllSubscriptions200JSONResponse{Subscriptions: ToSubscriptionsInfo(result)}, nil
}

func (h *Handler) Subscribe(ctx context.Context, request apiv1.SubscribeRequestObject) (apiv1.SubscribeResponseObject, error) {
	result, err := h.service.Add(ctx, dto.AddSubscriptionCommand{
		EndpointID: request.Id,
		EventTypes: request.Body.EventTypes,
	})

	if err != nil {
		return ToSubscribeErrorResponse(err), nil
	}

	return apiv1.Subscribe200JSONResponse{
		EndpointId:    request.Id,
		Subscriptions: ToSubscriptionsInfoFromDomain(result.Subscriptions),
	}, nil
}

func (h *Handler) DeleteSubscription(ctx context.Context, request apiv1.DeleteSubscriptionRequestObject) (apiv1.DeleteSubscriptionResponseObject, error) {
	if err := h.service.Delete(ctx, request.Id); err != nil {
		return ToDeleteErrorResponse(err), nil
	}

	return apiv1.DeleteSubscription204Response{}, nil
}
