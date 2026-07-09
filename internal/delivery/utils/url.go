package utils

import (
	"errors"
	"fmt"
	"net"
	"net/url"
)

func ValidateURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	if u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") {
		return errors.New("invalid URL")
	}

	host := u.Hostname()
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsLinkLocalMulticast() || ip.IsLinkLocalUnicast() {
			return errors.New("URL points to a private address: " + host)
		}
	}

	if host == "localhost" || host == "127.0.0.1" {
		return errors.New("localhost not allowed")
	}

	return nil
}
