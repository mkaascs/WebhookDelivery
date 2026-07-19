package utils

import (
	"context"
	"errors"
	"webhook-delivery/internal/domain"
)

func IsDomainError(err error) bool {
	return errors.Is(err, domain.ErrEndpointNotFound) || errors.Is(err, domain.ErrSubscriptionNotFound) ||
		errors.Is(err, domain.ErrSubscriptionAlreadyExists) || errors.Is(err, domain.ErrEventNotFound) ||
		errors.Is(err, domain.ErrEndpointIsInactive)
}

func IsCtxError(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
