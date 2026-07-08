package dto

import "time"

type RegisterEndpointResult struct {
	ID        string
	Secret    string
	IsActive  bool
	CreatedAt time.Time
}
