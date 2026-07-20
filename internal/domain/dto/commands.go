package dto

import (
	"encoding/json"
	"time"
	"webhook-delivery/internal/domain"
)

type RegisterEndpointCommand struct {
	URL         string
	EventTypes  []string
	Description *string
}

type GetAllEndpointsCommand struct {
	Page  int
	Limit int
}

type UpdateEndpointCommand struct {
	ID          string
	URL         *string
	IsActive    *bool
	Description *string
}

type AddSubscriptionCommand struct {
	EndpointID string
	EventTypes []string
}

type PublishEventCommand struct {
	Type    string
	Payload json.RawMessage
}

type UpdateDeliveryStatusCommand struct {
	ID               string
	Status           domain.DeliveryStatus
	Attempts         int
	NextRetryAt      time.Time
	LastError        *string
	LastResponseCode *int
}

type AddEndpointCommand struct {
	URL         string
	EventTypes  []string
	Description *string
	Secret      string
}
