package hmacauth

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpc"
)

// signOption returns an httpc.Option that signs every outgoing request.
// It buffers the body to compute the Digest header when needed, then
// restores the body so the HTTP transport can still read it.
func signOption(c Conf) httpc.Option {
	return func(r *http.Request) *http.Request {
		ct := r.Header.Get(headerContentType)
		isForm := strings.HasPrefix(ct, contentTypeForm) || strings.HasPrefix(ct, contentTypeMultipart)

		var bodyBytes []byte
		if r.Body != nil && r.Body != http.NoBody && !isForm {
			var err error
			bodyBytes, err = io.ReadAll(r.Body)
			if err != nil {
				logx.Errorf("failed to read request body for HMAC signing: %v", err)
				return r
			}
			r.Body.Close()
			r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		signRequest(r, c, bodyBytes, time.Now())
		return r
	}
}

// ClientOption is a functional option for configuring a CommonClient.
type (
	ClientOption func(*CommonClient)

	// Client is a Kong HMAC Auth client that automatically signs every outgoing
	// request using the configured algorithm and credentials.
	Client interface {
		// Do sends a signed request with structured data.
		// path must start with "/" (e.g. "/users").
		// data supports path, form, json, header tags (see go-zero httpc).
		// Pass nil for requests without a body.
		// The caller is responsible for closing resp.Body.
		Do(ctx context.Context, method, path string, data any) (*http.Response, error)

		// DoRaw sends a signed request with raw body bytes.
		// path must start with "/" (e.g. "/upload").
		// Use this for multipart uploads, binary payloads, or any custom body.
		// contentType must be set when body is non-empty.
		// Signing is handled by the same httpc.Option as Do.
		// The caller is responsible for closing resp.Body.
		DoRaw(ctx context.Context, method, path, contentType string, body []byte) (*http.Response, error)
	}

	// CommonClient is the concrete Client implementation.
	CommonClient struct {
		conf    Conf
		cli     *http.Client
		service httpc.Service
	}
)

// MustNewClient creates a new Kong HMAC Auth client and panics on error.
func MustNewClient(c Conf, opts ...ClientOption) Client {
	client, err := NewClient(c, opts...)
	logx.Must(err)
	return client
}

// NewClient creates a new Kong HMAC Auth client using http.DefaultClient.
// Timeout is controlled by the context passed to Do/DoRaw, consistent with
// go-zero httpc conventions.
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

	client.service = httpc.NewServiceWithClient(c.Host, client.cli, signOption(c))

	return client, nil
}

// WithClient returns a ClientOption that injects a custom *http.Client,
// giving the caller full control over connection pooling, timeouts, and TLS.
func WithClient(cli *http.Client) ClientOption {
	return func(c *CommonClient) {
		c.cli = cli
	}
}

// Do sends a signed request with structured data.
// path must start with "/" (e.g. "/users").
// data supports path, form, json, header tags (see go-zero httpc buildRequest).
// Pass nil for requests without a body.
func (c *CommonClient) Do(ctx context.Context, method, path string, data any) (*http.Response, error) {
	return c.service.Do(ctx, method, c.conf.Host+path, data)
}

// DoRaw sends a signed request with raw body bytes.
// path must start with "/" (e.g. "/upload").
// Signing is handled by the httpc.Option attached to the service,
// providing the same tracing, metrics, and circuit-breaker as Do.
func (c *CommonClient) DoRaw(ctx context.Context, method, path, contentType string, body []byte) (*http.Response, error) {
	if len(body) > 0 && contentType == "" {
		return nil, errors.New("contentType is required when body is present")
	}

	u := c.conf.Host + path

	var reader io.Reader
	if len(body) > 0 {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, reader)
	if err != nil {
		return nil, err
	}

	if len(body) > 0 {
		req.Header.Set(headerContentType, contentType)
	}

	return c.service.DoRequest(req)
}

// Parse reads the HTTP response and decodes it into val.
// It delegates to go-zero httpc.Parse, which supports response headers
// (via `header` tag) and JSON body (via `json` tag), and closes resp.Body
// automatically. For non-JSON responses (XML, binary, etc.), read resp.Body
// directly instead.
func Parse(resp *http.Response, err error, val any) error {
	if err != nil {
		return err
	}
	return httpc.Parse(resp, val)
}
