package dto

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
	URL         *string
	IsActive    *bool
	Description *string
}

type AddSubscriptionCommand struct {
	EndpointID string
	EventTypes []string
}
