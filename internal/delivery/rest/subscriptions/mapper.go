package subscriptions

import (
	"errors"
	apiv1 "webhook-delivery/gen/rest/v1"
	"webhook-delivery/internal/delivery/rest/common"
	"webhook-delivery/internal/domain"
)

var (
	errEndpointNotFound          = apiv1.Error{Code: new("endpoint_not_found"), Message: "endpoint not found"}
	errSubscriptionNotFound      = apiv1.Error{Code: new("subscription_not_found"), Message: "subscription not found"}
	errSubscriptionAlreadyExists = apiv1.Error{Code: new("subscription_already_exists"), Message: "subscription already exists"}
)

func ToGetAllErrorResponse(err error) apiv1.GetAllSubscriptionsResponseObject {
	switch {
	case errors.Is(err, domain.ErrEndpointNotFound), errors.Is(err, domain.ErrInvalidID):
		return apiv1.GetAllSubscriptions404JSONResponse(errEndpointNotFound)

	default:
		return apiv1.GetAllSubscriptions500JSONResponse(common.InternalError())
	}
}

func ToSubscribeErrorResponse(err error) apiv1.SubscribeResponseObject {
	switch {
	case errors.Is(err, domain.ErrEndpointNotFound), errors.Is(err, domain.ErrInvalidID):
		return apiv1.Subscribe404JSONResponse(errEndpointNotFound)

	case errors.Is(err, domain.ErrSubscriptionAlreadyExists):
		return apiv1.Subscribe409JSONResponse(errSubscriptionAlreadyExists)

	default:
		return apiv1.Subscribe500JSONResponse(common.InternalError())
	}
}

func ToDeleteErrorResponse(err error) apiv1.DeleteSubscriptionResponseObject {
	switch {
	case errors.Is(err, domain.ErrSubscriptionNotFound), errors.Is(err, domain.ErrInvalidID):
		return apiv1.DeleteSubscription404JSONResponse(errSubscriptionNotFound)

	default:
		return apiv1.DeleteSubscription500JSONResponse(common.InternalError())
	}
}

func ToSubscriptionsInfo(subscriptions []domain.Subscription) []apiv1.SubscriptionInfo {
	result := make([]apiv1.SubscriptionInfo, 0, len(subscriptions))
	for _, sub := range subscriptions {
		result = append(result, apiv1.SubscriptionInfo{
			Id:         domain.ToPublicSubscriptionID(sub.ID),
			EndpointId: domain.ToPublicEndpointID(sub.EndpointID),
			EventType:  sub.EventType,
			CreatedAt:  sub.CreatedAt,
		})
	}

	return result
}
