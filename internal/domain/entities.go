package domain

import (
	"encoding/json"
	"time"
)

type Endpoint struct {
	ID          string
	URL         string
	Secret      string
	EventTypes  []string
	Description *string
	IsActive    bool
	CreatedAt   time.Time
}

type Subscription struct {
	ID         string
	EndpointID string
	EventType  string
	CreatedAt  time.Time
}

type Event struct {
	ID        string
	Type      string
	Payload   json.RawMessage
	CreatedAt time.Time
}

type DeliveryStatus string

const (
	StatusDelivered = DeliveryStatus("delivered")
	StatusPending   = DeliveryStatus("pending")
	StatusFailed    = DeliveryStatus("failed")
)

type Deliveries struct {
	ID          string
	EndpointID  string
	EventID     string
	Status      DeliveryStatus
	Attempts    int
	MaxAttempts int
	NextRetryAt time.Time
	CreatedAt   time.Time
}
