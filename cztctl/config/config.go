package config

import (
	"errors"
	"strings"
)

const (
	// DefaultFormat defines a default naming style
	DefaultFormat = "gozero"
)

// Config defines the file naming style
type Config struct {
	NamingFormat string
}

// NewConfig creates an instance for Config
func NewConfig(format string) (*Config, error) {
	if len(format) == 0 {
		format = DefaultFormat
	}
	cfg := &Config{NamingFormat: format}
	if len(strings.TrimSpace(cfg.NamingFormat)) == 0 {
		return nil, errors.New("missing namingFormat")
	}
	return cfg, nil
}
