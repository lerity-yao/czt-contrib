package oss

import (
	"errors"
	"net/url"
	"strings"
)

// Conf holds the configuration for the OSS client.
type Conf struct {
	// Endpoint is the OSS region endpoint, e.g. "oss-cn-hangzhou.aliyuncs.com".
	// May include a scheme prefix ("http://" or "https://"); defaults to https.
	// Use "http://" for local S3-compatible services like MinIO.
	Endpoint string
	// Bucket is the OSS bucket name.
	Bucket string
	// AccessKeyId is the Alibaba Cloud AccessKey ID.
	AccessKeyId string
	// AccessKeySecret is the Alibaba Cloud AccessKey Secret.
	AccessKeySecret string
}

// Validate checks that required fields are present.
// It does not mutate any fields.
func (c Conf) Validate() error {
	if c.Endpoint == "" {
		return errors.New("endpoint is required")
	}
	if c.Bucket == "" {
		return errors.New("bucket is required")
	}
	if c.AccessKeyId == "" {
		return errors.New("access key id is required")
	}
	if c.AccessKeySecret == "" {
		return errors.New("access key secret is required")
	}
	return nil
}

// scheme extracts the URL scheme from Endpoint, defaulting to https.
func (c Conf) scheme() string {
	if strings.HasPrefix(c.Endpoint, "http://") {
		return "http"
	}
	return defaultScheme
}

// cleanEndpoint returns Endpoint with scheme prefix and trailing slashes removed.
func (c Conf) cleanEndpoint() string {
	ep := strings.TrimPrefix(c.Endpoint, "https://")
	ep = strings.TrimPrefix(ep, "http://")
	return strings.TrimRight(ep, "/")
}

// host returns the virtual-hosted style host: {bucket}.{endpoint}.
func (c Conf) host() string {
	return c.Bucket + "." + c.cleanEndpoint()
}

// objectURL builds the full request URL for the given object key.
// key may be empty for bucket-level operations (e.g. ListObjects).
func (c Conf) objectURL(key string) *url.URL {
	return &url.URL{
		Scheme: c.scheme(),
		Host:   c.host(),
		Path:   "/" + key,
	}
}
