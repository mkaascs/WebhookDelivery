package utils

import (
	"errors"
	"webhook-delivery/internal/domain"
)

func IsDomainError(err error) bool {
	return errors.Is(err, domain.ErrEndpointNotFound) || errors.Is(err, domain.ErrSubscriptionNotFound) ||
		errors.Is(err, domain.ErrSubscriptionAlreadyExists) || errors.Is(err, domain.ErrEventNotFount) ||
		errors.Is(err, domain.ErrEndpointIsInactive)
}
