package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type DeliveryStatus string

const (
	StatusDelivered  DeliveryStatus = "delivered"
	StatusPending    DeliveryStatus = "pending"
	StatusFailed     DeliveryStatus = "failed"
	StatusProcessing DeliveryStatus = "processing"
)

type Endpoint struct {
	ID          uuid.UUID
	URL         string
	Secret      string
	EventTypes  []string
	Description *string
	IsActive    bool
	CreatedAt   time.Time
}

type Subscription struct {
	ID         uuid.UUID
	EndpointID uuid.UUID
	EventType  string
	CreatedAt  time.Time
}

type Event struct {
	ID        uuid.UUID
	Type      string
	Payload   json.RawMessage
	CreatedAt time.Time
}

type Delivery struct {
	ID               uuid.UUID
	EndpointID       uuid.UUID
	EventID          uuid.UUID
	Status           DeliveryStatus
	Attempts         int
	MaxAttempts      int
	NextRetryAt      time.Time
	CreatedAt        time.Time
	LastResponseCode *int
	LastError        *string
}

type PendingDelivery struct {
	Delivery
	URL     string
	Secret  string
	Payload json.RawMessage
}
