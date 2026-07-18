package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

const secretSize = 32

func NewSecret() string {
	bytes := make([]byte, secretSize)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("whsec_%s", base64.StdEncoding.EncodeToString(bytes))
}
