package dto

import (
	"encoding/json"
	"time"
	"webhook-delivery/internal/domain"

	"github.com/google/uuid"
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
	ID          uuid.UUID
	URL         *string
	IsActive    *bool
	Description *string
}

type AddSubscriptionCommand struct {
	EndpointID uuid.UUID
	EventTypes []string
}

type PublishEventCommand struct {
	Type    string
	Payload json.RawMessage
}

type UpdateDeliveryStatusCommand struct {
	ID               uuid.UUID
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

type GetDeliveriesFromEventCommand struct {
	EventID uuid.UUID
	Page    int
	Limit   int
}
