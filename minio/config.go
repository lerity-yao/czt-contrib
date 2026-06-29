package minio

import "fmt"

// Conf holds the configuration for the MinIO client.
type Conf struct {
	// Endpoints is the list of MinIO server addresses for load balancing,
	// e.g. ["192.168.1.10:9000", "192.168.1.11:9000"].
	Endpoints []string
	// AccessKeyID is the access key for authentication.
	AccessKeyID string
	// SecretAccessKey is the secret key for authentication.
	SecretAccessKey string
	// UseSSL controls whether to use HTTPS (true) or HTTP (false).
	// Defaults to false since direct IP connections typically don't use SSL.
	UseSSL bool `json:",default=false"`
	// Region is the optional server region (e.g. "us-east-1").
	Region string `json:",optional"`
	// SignatureVersion selects the signing algorithm: "v2" or "v4".
	SignatureVersion string `json:",default=v4,options=v2|v4"`
	// SlowThreshold is the threshold in milliseconds for slow request logging.
	SlowThreshold int64 `json:",default=1000,optional"`
}

// Validate checks that required fields are present and values are valid.
func (c Conf) Validate() error {
	if len(c.Endpoints) == 0 {
		return fmt.Errorf("minio: at least one endpoint is required")
	}
	for i, ep := range c.Endpoints {
		if ep == "" {
			return fmt.Errorf("minio: endpoint at index %d is empty", i)
		}
	}
	if c.AccessKeyID == "" {
		return fmt.Errorf("minio: access key id is required")
	}
	if c.SecretAccessKey == "" {
		return fmt.Errorf("minio: secret access key is required")
	}
	if c.SignatureVersion != SignatureV2 && c.SignatureVersion != SignatureV4 {
		return fmt.Errorf("minio: signature version must be v2 or v4")
	}
	if c.SlowThreshold < 0 {
		return fmt.Errorf("minio: slow threshold must not be negative")
	}
	return nil
}
