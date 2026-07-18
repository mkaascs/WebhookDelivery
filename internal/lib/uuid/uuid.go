package uuid

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

const idLength = 16

func New(prefix string) string {
	bytes := make([]byte, idLength)
	_, _ = rand.Read(bytes)

	return fmt.Sprintf("%s_%s", prefix, base64.URLEncoding.EncodeToString(bytes))
}

func NewEndpoint() string {
	return New("ep")
}

func NewDelivery() string {
	return New("del")
}

func NewSubscription() string {
	return New("sub")
}

func NewEvent() string {
	return New("ev")
}
