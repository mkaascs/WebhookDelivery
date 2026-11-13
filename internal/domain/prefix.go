package domain

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const (
	PrefixEndpointID     = "ep"
	PrefixEventID        = "ev"
	PrefixDeliveryID     = "del"
	PrefixSubscriptionID = "sub"
)

func ToPublicDeliveryID(id uuid.UUID) string {
	return ToPublicID(PrefixDeliveryID, id)
}

func ToPublicEndpointID(id uuid.UUID) string {
	return ToPublicID(PrefixEndpointID, id)
}

func ToPublicEventID(id uuid.UUID) string {
	return ToPublicID(PrefixEventID, id)
}

func ToPublicSubscriptionID(id uuid.UUID) string {
	return ToPublicID(PrefixSubscriptionID, id)
}

func ToPublicID(prefix string, id uuid.UUID) string {
	return fmt.Sprintf("%s_%s", prefix, id.String())
}

func ParseDeliveryID(publicID string) (uuid.UUID, error) {
	return ParsePublicID(PrefixDeliveryID, publicID)
}

func ParseEndpointID(publicID string) (uuid.UUID, error) {
	return ParsePublicID(PrefixEndpointID, publicID)
}

func ParseEventID(publicID string) (uuid.UUID, error) {
	return ParsePublicID(PrefixEventID, publicID)
}

func ParseSubscriptionID(publicID string) (uuid.UUID, error) {
	return ParsePublicID(PrefixSubscriptionID, publicID)
}

func ParsePublicID(prefix, publicID string) (uuid.UUID, error) {
	raw, ok := strings.CutPrefix(publicID, prefix+"_")
	if !ok {
		return uuid.Nil, fmt.Errorf("%w: %q has no %q prefix", ErrInvalidID, publicID, prefix)
	}

	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: %w", ErrInvalidID, err)
	}

	return id, nil
}
