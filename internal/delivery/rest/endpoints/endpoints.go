package endpoints

import (
	"context"
	apiv1 "webhook-delivery/gen/rest/v1"
	"webhook-delivery/internal/delivery/rest/common"
	"webhook-delivery/internal/domain/dto"
	"webhook-delivery/internal/lib/ptr"
)

type Interface interface {
	GetAll(ctx context.Context, command dto.GetAllEndpointsCommand) (*dto.GetAllEndpointsResult, error)
	GetByID(ctx context.Context, id string) (*dto.GetEndpointResult, error)
	Register(ctx context.Context, command dto.RegisterEndpointCommand) (*dto.RegisterEndpointResult, error)
	Update(ctx context.Context, command dto.UpdateEndpointCommand) error
	Delete(ctx context.Context, id string) error
}

type Handler struct {
	service Interface
}

func NewHandler(service Interface) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetAllEndpoints(ctx context.Context, request apiv1.GetAllEndpointsRequestObject) (apiv1.GetAllEndpointsResponseObject, error) {
	page, limit := common.Pagination(request.Params.Page, request.Params.Limit)

	result, err := h.service.GetAll(ctx, dto.GetAllEndpointsCommand{
		Page:  page,
		Limit: limit,
	})

	if err != nil {
		return ToGetAllErrorResponse(err), nil
	}

	return apiv1.GetAllEndpoints200JSONResponse{
		Endpoints: ToEndpointsInfo(result.Endpoints),
		Page:      page,
		Limit:     limit,
		Total:     result.Total,
	}, nil
}

func (h *Handler) GetEndpoint(ctx context.Context, request apiv1.GetEndpointRequestObject) (apiv1.GetEndpointResponseObject, error) {
	result, err := h.service.GetByID(ctx, request.Id)
	if err != nil {
		return ToGetErrorResponse(err), nil
	}

	return apiv1.GetEndpoint200JSONResponse(ToEndpointInfo(*result)), nil
}

func (h *Handler) RegisterEndpoint(ctx context.Context, request apiv1.RegisterEndpointRequestObject) (apiv1.RegisterEndpointResponseObject, error) {
	if err := common.ValidateURL(request.Body.Url); err != nil {
		return apiv1.RegisterEndpoint400JSONResponse(common.InvalidURLError()), nil
	}

	eventTypes := common.GetDefaultArrayIfNull(ptr.Defer(request.Body.EventTypes))

	result, err := h.service.Register(ctx, dto.RegisterEndpointCommand{
		URL:         request.Body.Url,
		EventTypes:  eventTypes,
		Description: request.Body.Description,
	})

	if err != nil {
		return ToRegisterErrorResponse(err), nil
	}

	return apiv1.RegisterEndpoint200JSONResponse{
		Id:         result.ID,
		Url:        request.Body.Url,
		Secret:     result.Secret,
		IsActive:   result.IsActive,
		EventTypes: eventTypes,
		CreatedAt:  result.CreatedAt,
	}, nil
}

func (h *Handler) UpdateEndpoint(ctx context.Context, request apiv1.UpdateEndpointRequestObject) (apiv1.UpdateEndpointResponseObject, error) {
	if request.Body == nil {
		return apiv1.UpdateEndpoint204Response{}, nil
	}

	if request.Body.Url != nil {
		if err := common.ValidateURL(*request.Body.Url); err != nil {
			return apiv1.UpdateEndpoint400JSONResponse(common.InvalidURLError()), nil
		}
	}

	err := h.service.Update(ctx, dto.UpdateEndpointCommand{
		ID:          request.Id,
		URL:         request.Body.Url,
		IsActive:    request.Body.IsActive,
		Description: request.Body.Description,
	})

	if err != nil {
		return ToUpdateErrorResponse(err), nil
	}

	return apiv1.UpdateEndpoint204Response{}, nil
}

func (h *Handler) DeleteEndpoint(ctx context.Context, request apiv1.DeleteEndpointRequestObject) (apiv1.DeleteEndpointResponseObject, error) {
	if err := h.service.Delete(ctx, request.Id); err != nil {
		return ToDeleteErrorResponse(err), nil
	}

	return apiv1.DeleteEndpoint204Response{}, nil
}
