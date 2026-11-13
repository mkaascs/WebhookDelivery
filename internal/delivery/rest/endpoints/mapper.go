package endpoints

import (
	"errors"
	apiv1 "webhook-delivery/gen/rest/v1"
	"webhook-delivery/internal/delivery/rest/common"
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/lib/ptr"
)

var errEndpointNotFound = apiv1.Error{Code: new("endpoint_not_found"), Message: "endpoint not found"}

func ToGetAllErrorResponse(_ error) apiv1.GetAllEndpointsResponseObject {
	return apiv1.GetAllEndpoints500JSONResponse(common.InternalError())
}

func ToGetErrorResponse(err error) apiv1.GetEndpointResponseObject {
	switch {
	case errors.Is(err, domain.ErrEndpointNotFound), errors.Is(err, domain.ErrInvalidID):
		return apiv1.GetEndpoint404JSONResponse(errEndpointNotFound)

	default:
		return apiv1.GetEndpoint500JSONResponse(common.InternalError())
	}
}

func ToRegisterErrorResponse(_ error) apiv1.RegisterEndpointResponseObject {
	return apiv1.RegisterEndpoint500JSONResponse(common.InternalError())
}

func ToUpdateErrorResponse(err error) apiv1.UpdateEndpointResponseObject {
	switch {
	case errors.Is(err, domain.ErrEndpointNotFound), errors.Is(err, domain.ErrInvalidID):
		return apiv1.UpdateEndpoint404JSONResponse(errEndpointNotFound)

	default:
		return apiv1.UpdateEndpoint500JSONResponse(common.InternalError())
	}
}

func ToDeleteErrorResponse(err error) apiv1.DeleteEndpointResponseObject {
	switch {
	case errors.Is(err, domain.ErrEndpointNotFound), errors.Is(err, domain.ErrInvalidID):
		return apiv1.DeleteEndpoint404JSONResponse(errEndpointNotFound)

	default:
		return apiv1.DeleteEndpoint500JSONResponse(common.InternalError())
	}
}

func ToEndpointInfo(endpoint domain.Endpoint) apiv1.EndpointInfo {
	return apiv1.EndpointInfo{
		Id:          domain.ToPublicEndpointID(endpoint.ID),
		Url:         endpoint.URL,
		Description: ptr.Defer(endpoint.Description),
		IsActive:    endpoint.IsActive,
		EventTypes:  common.GetDefaultArrayIfNull(endpoint.EventTypes),
		CreatedAt:   endpoint.CreatedAt,
	}
}

func ToEndpointsInfo(endpoints []domain.Endpoint) []apiv1.EndpointInfo {
	result := make([]apiv1.EndpointInfo, 0, len(endpoints))
	for _, endpoint := range endpoints {
		result = append(result, ToEndpointInfo(endpoint))
	}

	return result
}
