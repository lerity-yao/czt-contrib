package consul

import (
	"errors"
	"fmt"
)

const (
	allEths       = "0.0.0.0"
	envPodIP      = "POD_IP"
	consulTags    = "consul_tags"
	CheckTypeTTL  = "ttl"
	CheckTypeHttp = "http"
	CheckTypeGrpc = "grpc"
	healthPort    = 6060
	healthPath    = "/healthz"
)

// CheckHttpConf is the http check config.
type CheckHttpConf struct {
	Method string `json:",default=GET,options=GET|POST"`
	Path   string `json:",default=/healthz"`
	Host   string `json:",default=0.0.0.0"`
	Port   int    `json:",default=6060"`
	Scheme string `json:",default=http,options=http|https"`
}

// CheckGrpcConf is the grpc check config.
type CheckGrpcConf struct {
	TLSServerName string `json:",optional"`
	TLSSkipVerify bool   `json:",default=true"`
	GRPCUseTLS    bool   `json:",default=false"`
}

// Conf is the config item with the given key on consul
// Host is the consul hosts. example: "localhost:8500"
// Key is the consul key. example: "service/name"
// Scheme is the consul scheme. example: "http"
// Tag is the consul tags. example: "tag1,tag2"
// Meta is the consul meta. example: {"key": "value"}
// TTL is the live check interval. example: 20
// ExpiredTTL is the deregistration time multiplier. example: 3
// CheckTimeout is the health check timeout. example: 3
// CheckType is the check type. example: "ttl", "http", "grpc"
// CheckHttp is the http check config.
// CheckGrpc is the grpc check config.
// CheckTypeTTL is the ttl check config.
type Conf struct {
	Host         string            // consul hosts
	Key          string            // consul key
	Scheme       string            `json:",default=http,options=http|https"`   // consul scheme
	Token        string            `json:",optional"`                          // consul token
	Tag          []string          `json:",optional"`                          // consul tags
	Meta         map[string]string `json:",optional"`                          // consul meta
	TTL          int               `json:",default=20"`                        // live check interval
	ExpiredTTL   int               `json:",default=3"`                         // Deregistration time multiplier. The actual deregistration time is calculated as TTL*ExpiredTTL in seconds.
	CheckTimeout int               `json:",default=3"`                         // health check timeout, http or grpc check timeout, ttl unuse
	CheckType    string            `json:",default=ttl,options=ttl|grpc|http"` // check type, ttl, http or grpc
	CheckHttp    CheckHttpConf
	CheckGrpc    CheckGrpcConf
}

// Validate validates c.
// Validate validates c.
func (c *Conf) Validate() error {
	if len(c.Host) == 0 {
		return errors.New("empty consul hosts")
	}
	if len(c.Key) == 0 {
		return errors.New("empty consul key")
	}

	if c.CheckType == "" {
		c.CheckType = CheckTypeTTL
	}
	if c.TTL == 0 {
		c.TTL = 20
	}

	if c.ExpiredTTL == 0 {
		c.ExpiredTTL = 3
	}

	if c.CheckTimeout == 0 {
		c.CheckTimeout = 3
	}

	if c.Scheme == "" {
		c.Scheme = "http"
	}

	switch c.CheckType {
	case CheckTypeTTL:
	case CheckTypeGrpc:
	case CheckTypeHttp:
		if c.CheckHttp.Scheme == "" {
			c.CheckHttp.Scheme = "http"
		}
		if c.CheckHttp.Method == "" {
			c.CheckHttp.Method = "GET"
		}
		if c.CheckHttp.Path == "" {
			c.CheckHttp.Path = healthPath
		}
		if c.CheckHttp.Port == 0 {
			c.CheckHttp.Port = healthPort
		}
		if c.CheckHttp.Host == "" {
			c.CheckHttp.Host = fmt.Sprintf("0.0.0.0")
		}
	default:
		return fmt.Errorf("unknown check type: %s", c.CheckType)

	}

	return nil
}
