package deliveries

import (
	"errors"
	apiv1 "webhook-delivery/gen/rest/v1"
	"webhook-delivery/internal/delivery/rest/common"
	"webhook-delivery/internal/domain"
)

var errDeliveryNotFound = apiv1.Error{Code: new("delivery_not_found"), Message: "delivery not found"}

func ToGetErrorResponse(err error) apiv1.GetDeliveryResponseObject {
	switch {
	case errors.Is(err, domain.ErrDeliveryNotFound):
		return apiv1.GetDelivery404JSONResponse(errDeliveryNotFound)

	default:
		return apiv1.GetDelivery500JSONResponse(common.InternalError())
	}
}

func ToGetEventDeliveriesErrorResponse(_ error) apiv1.GetEventDeliveriesResponseObject {
	return apiv1.GetEventDeliveries500JSONResponse(common.InternalError())
}

func ToRetryErrorResponse(err error) apiv1.RetryDeliveryResponseObject {
	switch {
	case errors.Is(err, domain.ErrDeliveryNotFound):
		return apiv1.RetryDelivery404JSONResponse(errDeliveryNotFound)

	default:
		return apiv1.RetryDelivery500JSONResponse(common.InternalError())
	}
}

func ToDeliveryInfo(delivery domain.Delivery) apiv1.DeliveryInfo {
	return apiv1.DeliveryInfo{
		Id:               delivery.ID,
		EndpointId:       delivery.EndpointID,
		EventId:          delivery.EventID,
		Status:           apiv1.DeliveryInfoStatus(delivery.Status),
		Attempts:         delivery.Attempts,
		MaxAttempts:      delivery.MaxAttempts,
		NextRetryAt:      delivery.NextRetryAt,
		CreatedAt:        delivery.CreatedAt,
		LastResponseCode: delivery.LastResponseCode,
		LastError:        delivery.LastError,
	}
}

func ToDeliveriesInfo(deliveries []domain.Delivery) []apiv1.DeliveryInfo {
	result := make([]apiv1.DeliveryInfo, 0, len(deliveries))
	for _, delivery := range deliveries {
		result = append(result, ToDeliveryInfo(delivery))
	}

	return result
}
