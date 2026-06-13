package gateway

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpc"
)

// Client is an API Gateway client that automatically signs every request
// using the Alibaba Cloud API Gateway v1 HMAC-SHA256 algorithm.
type (
	ClientOption func(*CommonClient)

	Client interface {
		// Do sends a signed request with structured data.
		// data supports path, form, json, header tags (see go-zero httpc).
		// Pass nil for requests without body.
		Do(ctx context.Context, method, path string, data any) (*http.Response, error)

		// DoRaw sends a signed request with raw body bytes.
		// Use this for multipart uploads or custom bodies.
		DoRaw(ctx context.Context, method, path, contentType string, body []byte) (*http.Response, error)
	}

	CommonClient struct {
		conf    Conf
		cli     *http.Client
		service httpc.Service
	}
)

func MustNewClient(c Conf, opts ...ClientOption) Client {
	client, err := NewClient(c, opts...)
	logx.Must(err)
	return client
}

// NewClient creates a new API Gateway client with http.DefaultClient.
// Timeout is controlled by the context passed to Do/DoRaw, consistent with go-zero httpc conventions.
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

	client.service = httpc.NewServiceWithClient(c.Host, client.cli, signOption(client.conf))

	return client, nil
}

// Do sends a signed request with structured data.
// data supports path, form, json, header tags (see go-zero httpc buildRequest).
// Pass nil for requests without body.
// The caller is responsible for closing resp.Body.
func (c *CommonClient) Do(ctx context.Context, method, path string, data any) (*http.Response, error) {
	return c.service.Do(ctx, method, c.conf.Host+path, data)
}

// DoRaw sends a signed request with raw body bytes.
// Use this for multipart uploads or custom bodies.
// The caller is responsible for closing resp.Body.
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

func WithClient(cli *http.Client) ClientOption {
	return func(c *CommonClient) {
		c.cli = cli
	}
}

// Parse reads the HTTP response and decodes it into val.
// It delegates to go-zero httpc.Parse, which supports response headers
// (via `header` tag) and JSON body (via `json` tag), and closes resp.Body automatically.
// For non-JSON responses (XML, binary, etc.), read resp.Body directly instead.
func Parse(resp *http.Response, err error, val any) error {
	if err != nil {
		return err
	}
	return httpc.Parse(resp, val)
}

// signOption returns an httpc.Option that signs every outgoing request.
func signOption(c Conf) httpc.Option {
	appSecret := []byte(c.AppSecret)
	return func(r *http.Request) *http.Request {
		ct := r.Header.Get(headerContentType)
		skipMD5 := strings.HasPrefix(ct, contentTypeForm) || strings.HasPrefix(ct, contentTypeMultipart)

		var bodyBytes []byte
		if r.Body != nil && r.Body != http.NoBody {
			if !skipMD5 || r.GetBody == nil {
				// Must buffer body: either for MD5 computation, or body is not re-creatable.
				bodyBytes, _ = io.ReadAll(r.Body)
				r.Body.Close()
				if r.GetBody != nil {
					r.Body, _ = r.GetBody()
				} else {
					r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
				}
			}
			// else: form/multipart with re-creatable body — skip buffering entirely.
		}
		signRequest(r, c.AppKey, appSecret, bodyBytes)
		return r
	}
}
