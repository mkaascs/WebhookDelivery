package domain

import "errors"

var (
	ErrEndpointNotFound   = errors.New("endpoint not found")
	ErrEventNotFount      = errors.New("event not found")
	ErrEndpointIsInactive = errors.New("endpoint is inactive")
)
