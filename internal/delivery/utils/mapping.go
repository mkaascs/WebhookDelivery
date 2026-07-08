package utils

import (
	"webhook-delivery/internal/delivery/handlers/endpoints"
	"webhook-delivery/internal/domain/dto"
)

func ToEndpointInfo(result dto.GetEndpointResult) endpoints.EndpointInfo {
	return endpoints.EndpointInfo{
		ID:          result.ID,
		URL:         result.URL,
		EventTypes:  result.EventTypes,
		Description: result.Description,
		IsActive:    result.IsActive,
		CreatedAt:   result.CreatedAt,
	}
}
