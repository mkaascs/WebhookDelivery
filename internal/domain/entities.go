package domain

import (
	"encoding/json"
	"time"
)

type Endpoint struct {
	ID          string
	URL         string
	Secret      []byte
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
	ID        int64
	Type      string
	Payload   json.RawMessage
	CreatedAt time.Time
}
