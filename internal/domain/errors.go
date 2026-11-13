package domain

import (
	"context"
	"errors"
)

var (
	ErrEndpointNotFound          = errors.New("endpoint not found")
	ErrSubscriptionNotFound      = errors.New("subscription not found")
	ErrSubscriptionAlreadyExists = errors.New("subscription already exists")
	ErrEventNotFound             = errors.New("event not found")
	ErrEndpointIsInactive        = errors.New("endpoint is inactive")
	ErrDeliveryNotFound          = errors.New("delivery not found")
	ErrInvalidID                 = errors.New("invalid id")
)

func IsDomainError(err error) bool {
	return errors.Is(err, ErrEndpointNotFound) || errors.Is(err, ErrSubscriptionNotFound) ||
		errors.Is(err, ErrSubscriptionAlreadyExists) || errors.Is(err, ErrEventNotFound) ||
		errors.Is(err, ErrEndpointIsInactive) || errors.Is(err, ErrDeliveryNotFound)
}

func IsCtxError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
