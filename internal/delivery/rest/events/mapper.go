package events

import (
	"errors"
	apiv1 "webhook-delivery/gen/rest/v1"
	"webhook-delivery/internal/delivery/rest/common"
	"webhook-delivery/internal/domain"
)

var errEventNotFound = apiv1.Error{Code: new("event_not_found"), Message: "event not found"}

func ToPublishErrorResponse(_ error) apiv1.PublishEventResponseObject {
	return apiv1.PublishEvent500JSONResponse(common.InternalError())
}

func ToGetErrorResponse(err error) apiv1.GetEventResponseObject {
	switch {
	case errors.Is(err, domain.ErrEventNotFound):
		return apiv1.GetEvent404JSONResponse(errEventNotFound)

	default:
		return apiv1.GetEvent500JSONResponse(common.InternalError())
	}
}
