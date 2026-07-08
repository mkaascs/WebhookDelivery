package endpoints

import (
	"time"
	"webhook-delivery/internal/delivery/utils"
)

type EndpointInfo struct {
	ID          string
	URL         string
	EventTypes  string
	Description *string
	IsActive    bool
	CreatedAt   time.Time
}

type GetResponse struct {
	utils.Response
	EndpointInfo
}
