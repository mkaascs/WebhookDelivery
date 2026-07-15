package dto

import (
	"encoding/json"
	"time"
	"webhook-delivery/internal/domain"
)

type RegisterEndpointResult struct {
	ID        string
	Secret    string
	IsActive  bool
	CreatedAt time.Time
}

type GetEndpointResult struct {
	ID          string
	URL         string
	EventTypes  []string
	Description *string
	IsActive    bool
	CreatedAt   time.Time
}

type GetAllEndpointsResult struct {
	Total     int
	Endpoints []GetEndpointResult
}

type AddSubscriptionResult struct {
	Subscriptions []domain.Subscription
}

type PublishEventResult struct {
	Event             domain.Event
	DeliveriesCreated int
}

type GetEventResult struct {
	Event domain.Event
}

type ClaimPendingResult struct {
	URL         string
	Secret      []byte
	Payload     json.RawMessage
	Attempts    int
	MaxAttempts int
	NextRetryAt time.Time
}
