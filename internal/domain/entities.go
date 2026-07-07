package domain

import (
	"encoding/json"
	"time"
)

type Endpoint struct {
	ID        int64
	URL       string
	Secret    []byte
	IsActive  bool
	CreatedAt time.Time
}

type Subscription struct {
	ID         int64
	EndpointID int64
	EventType  string
	CreatedAt  time.Time
}

type Event struct {
	ID        int64
	Type      string
	Payload   json.RawMessage
	CreatedAt time.Time
}
