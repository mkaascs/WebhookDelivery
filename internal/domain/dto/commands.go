package dto

import "encoding/json"

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
