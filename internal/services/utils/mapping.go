package utils

import (
	"webhook-delivery/internal/domain"
	"webhook-delivery/internal/domain/dto"
)

func EndpointToResult(endpoint domain.Endpoint) dto.GetEndpointResult {
	return dto.GetEndpointResult{
		ID:          endpoint.ID,
		URL:         endpoint.URL,
		EventTypes:  endpoint.EventTypes,
		Description: endpoint.Description,
		IsActive:    endpoint.IsActive,
		CreatedAt:   endpoint.CreatedAt,
	}
}
