package gateway

import (
	"errors"
	"strings"
)

// Conf Config holds the configuration for the API Gateway client.
type Conf struct {
	// Host is the gateway endpoint, must start with "http://" or "https://".
	Host string
	// AppKey is the AppKey assigned to the API Gateway app.
	AppKey string
	// AppSecret is the AppSecret used for HMAC-SHA256 signing.
	AppSecret string
}

// Validate checks that required fields are present and valid.
func (c *Conf) Validate() error {
	if c.Host == "" {
		return errors.New("host is required")
	}

	if !strings.HasPrefix(c.Host, "https://") && !strings.HasPrefix(c.Host, "http://") {
		return errors.New("host must start with http:// or https://")
	}

	// Trim trailing slashes to avoid double-slash in Host+path concatenation.
	c.Host = strings.TrimRight(c.Host, "/")

	if c.AppKey == "" {
		return errors.New("app key is required")
	}

	if c.AppSecret == "" {
		return errors.New("app secret is required")
	}
	return nil
}
