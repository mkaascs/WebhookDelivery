package dto

import "time"

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
