package subscriptions

import (
	"context"
	apiv1 "webhook-delivery/gen/rest/v1"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"

	"github.com/google/uuid"
)

type Interface interface {
	GetAll(ctx context.Context, endpointID uuid.UUID) ([]domain.Subscription, error)
	Add(ctx context.Context, command dto.AddSubscriptionCommand) ([]domain.Subscription, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type Handler struct {
	service Interface
}

func NewHandler(service Interface) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetAllSubscriptions(ctx context.Context, request apiv1.GetAllSubscriptionsRequestObject) (apiv1.GetAllSubscriptionsResponseObject, error) {
	endpointID, err := domain.ParseEndpointID(request.Id)
	if err != nil {
		return ToGetAllErrorResponse(err), nil
	}

	subscriptions, err := h.service.GetAll(ctx, endpointID)
	if err != nil {
		return ToGetAllErrorResponse(err), nil
	}

	return apiv1.GetAllSubscriptions200JSONResponse{Subscriptions: ToSubscriptionsInfo(subscriptions)}, nil
}

func (h *Handler) Subscribe(ctx context.Context, request apiv1.SubscribeRequestObject) (apiv1.SubscribeResponseObject, error) {
	endpointID, err := domain.ParseEndpointID(request.Id)
	if err != nil {
		return ToSubscribeErrorResponse(err), nil
	}

	subscriptions, err := h.service.Add(ctx, dto.AddSubscriptionCommand{
		EndpointID: endpointID,
		EventTypes: request.Body.EventTypes,
	})

	if err != nil {
		return ToSubscribeErrorResponse(err), nil
	}

	return apiv1.Subscribe200JSONResponse{
		EndpointId:    domain.ToPublicEndpointID(endpointID),
		Subscriptions: ToSubscriptionsInfo(subscriptions),
	}, nil
}

func (h *Handler) DeleteSubscription(ctx context.Context, request apiv1.DeleteSubscriptionRequestObject) (apiv1.DeleteSubscriptionResponseObject, error) {
	id, err := domain.ParseSubscriptionID(request.Id)
	if err != nil {
		return ToDeleteErrorResponse(err), nil
	}

	if err := h.service.Delete(ctx, id); err != nil {
		return ToDeleteErrorResponse(err), nil
	}

	return apiv1.DeleteSubscription204Response{}, nil
}
