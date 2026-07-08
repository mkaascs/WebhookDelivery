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
