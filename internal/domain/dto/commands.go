package dto

type RegisterEndpointCommand struct {
	URL         string
	EventTypes  []string
	Description *string
}
