package domain

import "errors"

var (
	ErrEndpointNotFound          = errors.New("endpoint not found")
	ErrSubscriptionNotFound      = errors.New("subscription not found")
	ErrSubscriptionAlreadyExists = errors.New("subscription already exists")
	ErrEventNotFount             = errors.New("event not found")
	ErrEndpointIsInactive        = errors.New("endpoint is inactive")
)
