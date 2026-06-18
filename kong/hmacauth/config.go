package hmacauth

import (
	"errors"
	"fmt"
	"strings"
)

// Conf holds the configuration for the Kong HMAC Auth client.
type Conf struct {
	// Host is the Kong gateway endpoint, must start with "http://" or "https://".
	Host string
	// Username is the Kong consumer credential username (key id).
	Username string
	// Secret is the shared secret used for HMAC signing.
	Secret string
	// Algorithm is the HMAC signing algorithm.
	// Supported values: "hmac-sha1", "hmac-sha224", "hmac-sha256", "hmac-sha384", "hmac-sha512".
	// Defaults to "hmac-sha256" if omitted.
	Algorithm string `json:",optional,default=hmac-sha256"`
	// Headers is the ordered list of headers to include in the signing string.
	// Supports standard HTTP headers (lowercase, e.g. "date", "host", "digest")
	// and the pseudo-header:
	//   @request-target (lowercase method + path + query)
	// Defaults to ["date", "@request-target"] if omitted.
	Headers []string `json:",optional"`
}

// Validate checks that all required fields are present and valid.
// It also applies defaults for Algorithm and Headers, and normalises Host.
func (c *Conf) Validate() error {
	if c.Host == "" {
		return errors.New("host is required")
	}
	if !strings.HasPrefix(c.Host, "https://") && !strings.HasPrefix(c.Host, "http://") {
		return errors.New("host must start with http:// or https://")
	}
	// Trim trailing slashes to avoid double-slash in Host+path concatenation.
	c.Host = strings.TrimRight(c.Host, "/")

	if c.Username == "" {
		return errors.New("username is required")
	}
	if c.Secret == "" {
		return errors.New("secret is required")
	}

	// Apply default algorithm.
	if c.Algorithm == "" {
		c.Algorithm = defaultAlgorithm
	} else {
		c.Algorithm = strings.ToLower(c.Algorithm)
		if !validAlgorithms[c.Algorithm] {
			return fmt.Errorf("unsupported algorithm %q, must be one of: hmac-sha1, hmac-sha224, hmac-sha256, hmac-sha384, hmac-sha512", c.Algorithm)
		}
	}

	// Apply default headers.
	if len(c.Headers) == 0 {
		c.Headers = defaultSignHeaders()
	} else {
		// Normalise to lowercase.
		for i, h := range c.Headers {
			c.Headers[i] = strings.ToLower(h)
		}
	}

	return nil
}
