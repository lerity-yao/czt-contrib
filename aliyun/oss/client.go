package oss

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpc"
)

// Option customizes request headers for PutObject / GetObject.
type Option func(map[string]string)

// ListOption customizes ListObjects parameters.
type ListOption func(*listConfig)

// ClientOption customizes the CommonClient.
type ClientOption func(*CommonClient)

type (
	// Client is an Alibaba Cloud OSS client that automatically signs
	// every request using the OSS V1 HMAC-SHA1 algorithm.
	Client interface {
		// Do sends a signed request with raw body.
		// key is the object key (may be empty for bucket-level operations).
		// headers can include Content-Type, x-oss-* headers, etc.
		// body is the raw request body, nil for no body.
		// The caller is responsible for closing resp.Body.
		Do(ctx context.Context, method, key string, headers map[string]string, body []byte) (*http.Response, error)

		// PutObject uploads an object.
		PutObject(ctx context.Context, key string, body []byte, opts ...Option) error

		// GetObject downloads an object.
		// Returns a ReadCloser that the caller must close.
		GetObject(ctx context.Context, key string, opts ...Option) (io.ReadCloser, error)

		// DeleteObject deletes an object.
		DeleteObject(ctx context.Context, key string) error

		// HeadObject retrieves object metadata without downloading the body.
		HeadObject(ctx context.Context, key string) (*ObjectMeta, error)

		// CopyObject copies an object within the same bucket.
		CopyObject(ctx context.Context, destKey, srcKey string) error

		// ListObjects lists objects in the bucket.
		ListObjects(ctx context.Context, opts ...ListOption) (*ListBucketResult, error)
	}

	// CommonClient implements Client using go-zero httpc for circuit breaking.
	CommonClient struct {
		conf    Conf
		cli     *http.Client
		service httpc.Service
	}
)

// listConfig holds parameters for ListObjects.
type listConfig struct {
	prefix       string
	marker       string
	maxKeys      int
	delimiter    string
	encodingType string
}

// MustNewClient creates a Client and panics on validation error.
func MustNewClient(c Conf, opts ...ClientOption) Client {
	client, err := NewClient(c, opts...)
	logx.Must(err)
	return client
}

// NewClient creates a new OSS client with http.DefaultClient.
// Timeout is controlled by the context passed to each method,
// consistent with go-zero httpc conventions.
func NewClient(c Conf, opts ...ClientOption) (Client, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	client := &CommonClient{conf: c}
	for _, opt := range opts {
		opt(client)
	}

	if client.cli == nil {
		client.cli = http.DefaultClient
	}

	client.service = httpc.NewServiceWithClient(c.host(), client.cli, signOption(c))

	return client, nil
}

// WithClient injects a custom *http.Client (TLS, connection pool, etc.).
func WithClient(cli *http.Client) ClientOption {
	return func(c *CommonClient) {
		c.cli = cli
	}
}

// Do sends a signed request with raw body.
// The caller is responsible for closing resp.Body.
func (c *CommonClient) Do(ctx context.Context, method, key string, headers map[string]string, body []byte) (*http.Response, error) {
	u := c.conf.objectURL(key)
	return c.doRequest(ctx, method, u, headers, body)
}

// PutObject uploads an object to the specified key.
func (c *CommonClient) PutObject(ctx context.Context, key string, body []byte, opts ...Option) error {
	headers := make(map[string]string)
	for _, opt := range opts {
		opt(headers)
	}

	resp, err := c.Do(ctx, http.MethodPut, key, headers, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkStatus(resp)
}

// GetObject downloads an object.
// Returns a ReadCloser that the caller must close.
func (c *CommonClient) GetObject(ctx context.Context, key string, opts ...Option) (io.ReadCloser, error) {
	headers := make(map[string]string)
	for _, opt := range opts {
		opt(headers)
	}

	resp, err := c.Do(ctx, http.MethodGet, key, headers, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		return nil, parseError(resp)
	}

	return resp.Body, nil
}

// DeleteObject deletes an object.
func (c *CommonClient) DeleteObject(ctx context.Context, key string) error {
	resp, err := c.Do(ctx, http.MethodDelete, key, nil, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkStatus(resp)
}

// HeadObject retrieves object metadata without downloading the body.
func (c *CommonClient) HeadObject(ctx context.Context, key string) (*ObjectMeta, error) {
	resp, err := c.Do(ctx, http.MethodHead, key, nil, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// HEAD responses have empty body; use status code for error detection.
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("oss: HTTP %d", resp.StatusCode)
	}

	meta := &ObjectMeta{
		Size:         resp.ContentLength,
		ContentType:  resp.Header.Get(headerContentType),
		ETag:         resp.Header.Get("ETag"),
		LastModified: resp.Header.Get("Last-Modified"),
		Metadata:     make(map[string]string),
	}

	for k, v := range resp.Header {
		lk := strings.ToLower(k)
		if strings.HasPrefix(lk, "x-oss-meta-") {
			meta.Metadata[lk] = strings.Join(v, ",")
		}
	}

	return meta, nil
}

// CopyObject copies an object within the same bucket.
func (c *CommonClient) CopyObject(ctx context.Context, destKey, srcKey string) error {
	headers := map[string]string{
		headerCopySource: "/" + c.conf.Bucket + "/" + srcKey,
	}

	resp, err := c.Do(ctx, http.MethodPut, destKey, headers, nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return checkStatus(resp)
}

// ListObjects lists objects in the bucket.
func (c *CommonClient) ListObjects(ctx context.Context, opts ...ListOption) (*ListBucketResult, error) {
	cfg := listConfig{}
	for _, opt := range opts {
		opt(&cfg)
	}

	u := c.conf.objectURL("")
	q := u.Query()
	if cfg.prefix != "" {
		q.Set("prefix", cfg.prefix)
	}
	if cfg.marker != "" {
		q.Set("marker", cfg.marker)
	}
	if cfg.maxKeys > 0 {
		q.Set("max-keys", strconv.Itoa(cfg.maxKeys))
	}
	if cfg.delimiter != "" {
		q.Set("delimiter", cfg.delimiter)
	}
	if cfg.encodingType != "" {
		q.Set("encoding-type", cfg.encodingType)
	}
	u.RawQuery = q.Encode()

	resp, err := c.doRequest(ctx, http.MethodGet, u, nil, nil)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, parseError(resp)
	}

	var result ListBucketResult
	if err := Parse(resp, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Parse decodes the XML response body into val and closes the body.
func Parse(resp *http.Response, err error, val any) error {
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return xml.NewDecoder(resp.Body).Decode(val)
}

// --- Option helpers ---

// WithContentType sets the Content-Type header.
func WithContentType(ct string) Option {
	return func(h map[string]string) {
		h[headerContentType] = ct
	}
}

// WithMeta sets a custom metadata header (x-oss-meta-{key}).
func WithMeta(key, value string) Option {
	return func(h map[string]string) {
		h["x-oss-meta-"+key] = value
	}
}

// WithHeader sets an arbitrary request header.
func WithHeader(key, value string) Option {
	return func(h map[string]string) {
		h[key] = value
	}
}

// --- ListOption helpers ---

// WithPrefix limits ListObjects to keys beginning with the given prefix.
func WithPrefix(prefix string) ListOption {
	return func(c *listConfig) {
		c.prefix = prefix
	}
}

// WithMarker sets the starting key for ListObjects pagination.
func WithMarker(marker string) ListOption {
	return func(c *listConfig) {
		c.marker = marker
	}
}

// WithMaxKeys sets the maximum number of keys to return (1–1000).
func WithMaxKeys(n int) ListOption {
	return func(c *listConfig) {
		c.maxKeys = n
	}
}

// WithDelimiter sets the delimiter for grouping keys in ListObjects.
func WithDelimiter(d string) ListOption {
	return func(c *listConfig) {
		c.delimiter = d
	}
}

// WithEncodingType sets the encoding type for object keys in the response (e.g. "url").
func WithEncodingType(t string) ListOption {
	return func(c *listConfig) {
		c.encodingType = t
	}
}

// --- internal helpers ---

// doRequest builds and sends an HTTP request through the signed service.
func (c *CommonClient) doRequest(ctx context.Context, method string, u *url.URL, headers map[string]string, body []byte) (*http.Response, error) {
	var reader io.Reader
	if len(body) > 0 {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reader)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return c.service.DoRequest(req)
}

// checkStatus reads the response and returns an error if the status code indicates failure.
func checkStatus(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return nil
	}
	return parseError(resp)
}

// parseError decodes an OSS error XML response into a ServiceError.
func parseError(resp *http.Response) error {
	var svcErr ServiceError
	if err := xml.NewDecoder(resp.Body).Decode(&svcErr); err != nil {
		resp.Body.Close()
		return fmt.Errorf("oss: HTTP %d", resp.StatusCode)
	}
	resp.Body.Close()

	if svcErr.Code == "" {
		return fmt.Errorf("oss: HTTP %d", resp.StatusCode)
	}
	svcErr.StatusCode = resp.StatusCode
	return &svcErr
}
