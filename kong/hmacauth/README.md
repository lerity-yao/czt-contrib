# hmacauth

[中文](./readme-cn.md)

A Go client for Kong HMAC Auth built on [go-zero](https://github.com/zeromicro/go-zero) httpc, with automatic **HMAC signing** and built-in **circuit breaker**.

Follows the official [Kong HMAC Auth plugin](https://developer.konghq.com/plugins/hmac-auth/) specification (based on [draft-cavage-http-signatures](https://tools.ietf.org/html/draft-cavage-http-signatures) draft, using the `@request-target` pseudo-header).

## Features

- 🔐 **Auto Signing** — Automatically injects `Authorization` signature header (`hmac username="...", algorithm="...", headers="...", signature="..."`) into every request, transparent to the caller
- 🔀 **Dual-method API** — `Do` for structured data (JSON / form / path / header auto-mapping), `DoRaw` for raw bytes (file uploads, XML, plain text, etc.)
- 🛡️ **Circuit Breaker** — Built on go-zero httpc, automatically shares circuit breakers per Host, trips when error rate is too high
- 🔧 **Custom HTTP Client** — Inject a custom `*http.Client` via `WithClient` (timeout, TLS, connection pool, etc.); defaults to `http.DefaultClient` if not provided
- 🧩 **Multi-algorithm Support** — Full coverage of five algorithms: `hmac-sha1`, `hmac-sha224`, `hmac-sha256`, `hmac-sha384`, `hmac-sha512`
- 📋 **Flexible Signing Headers** — Supports custom header lists for signing; defaults to `["date", "@request-target"]`
- 📦 **Body Integrity** — When `Headers` includes `digest`, automatically computes `Digest: SHA-256=...` header (including empty body scenarios)

## Installation

```bash
go get github.com/lerity-yao/czt-contrib/kong/hmacauth@v0.0.1
```

## Configuration

### Conf

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `Host` | string | Yes | — | Kong gateway address, must start with `http://` or `https://` |
| `Username` | string | Yes | — | Kong consumer credential username (i.e., key id) |
| `Secret` | string | Yes | — | HMAC signing secret |
| `Algorithm` | string | No | `hmac-sha256` | Signing algorithm; supports `hmac-sha1` / `hmac-sha224` / `hmac-sha256` / `hmac-sha384` / `hmac-sha512` |
| `Headers` | []string | No | `["date", "@request-target"]` | Header list to include in signing (lowercase); supports standard HTTP headers and `@request-target` pseudo-header |

Calling `NewClient` automatically runs `Validate()` to verify the above fields and fills in default values for `Algorithm` and `Headers`.

> `Algorithm` and `Headers` support go-zero `conf.MustLoad` `optional` / `default` tags; they can be omitted in YAML configuration files.

## API Reference

### Constructors

| Function | Signature | Description |
|----------|-----------|-------------|
| `MustNewClient` | `func MustNewClient(c Conf, opts ...ClientOption) Client` | Create Client, panic on validation failure |
| `NewClient` | `func NewClient(c Conf, opts ...ClientOption) (Client, error)` | Create Client, return error on validation failure |

### ClientOption

| Option | Parameter | Description |
|--------|-----------|-------------|
| `WithClient` | `*http.Client` | Inject a custom HTTP Client (timeout, TLS, connection pool, etc.); defaults to `http.DefaultClient` if not provided |

### Client Interface Methods

| Method | Signature | Use Case |
|--------|-----------|----------|
| `Do` | `Do(ctx, method, path string, data any) (*http.Response, error)` | Structured requests: JSON body / form query / path params / header auto-mapping |
| `DoRaw` | `DoRaw(ctx, method, path, contentType string, body []byte) (*http.Response, error)` | Raw requests: file uploads, XML, plain text, encrypted payloads, etc. |

> **Note**: `DoRaw` **requires** a non-empty `contentType` when `body` is non-empty, otherwise it returns an error. For requests without a body (e.g., GET), `contentType` can be empty.

### Response Parsing

| Function | Signature | Use Case |
|----------|-----------|----------|
| `Parse` | `Parse(resp *http.Response, err error, val any) error` | JSON responses: parses header + body, automatically closes Body |

`Parse` delegates to go-zero `httpc.Parse` internally, supporting response header (`header` tag) and JSON body (`json` tag) parsing. For non-JSON responses (XML, binary, etc.), read `resp.Body` directly.

> Use standard library `http.MethodGet`, `http.MethodPost`, etc. for HTTP methods.

## Advanced Guide

### Signing Mechanism

The SDK follows the official [Kong HMAC Auth plugin](https://developer.konghq.com/plugins/hmac-auth/) specification. Each request automatically performs the following:

1. **Set default headers** (when not already set): `Date` (GMT format), `User-Agent`, `Host` (when host is in the signing list)
2. **Calculate Digest** (when `Headers` includes `digest`): Compute `SHA-256` of non-form / non-multipart body, formatted as `Digest: SHA-256=<base64>`; empty body also computes a zero-length digest
3. **Build the string to sign**: Concatenate in `Headers` list order, each line formatted as `header-name: value`, separated by `\n`, with no trailing newline on the last line

```
date: Thu, 22 Jun 2023 17:15:21 GMT
@request-target: get /v1/users?page=1
host: api.example.com
digest: SHA-256=47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=
```

4. **Calculate signature**: HMAC the string to sign with `Secret` (algorithm specified by `Algorithm`), Base64-encode the result
5. **Set Authorization header**:

```
Authorization: hmac username="alice", algorithm="hmac-sha256", headers="date @request-target", signature="abc123..."
```

### @request-target Pseudo-header

Kong HMAC Auth uses `@request-target` as a pseudo-header to represent the request target line, with the value `method /path?query` (method in lowercase):

```
@request-target: post /v1/users?name=tom
```

> Kong officially recommends using `@request-target` instead of the legacy `request-line`.

### Signing Header List Configuration

`Headers` determines which headers participate in signing. By default, only `date` and `@request-target` are signed. Kong's recommended strong signing configuration is `@request-target` + `host` + `date`:

```go
hmacauth.Conf{
    Host:     "https://api.example.com",
    Username: "alice",
    Secret:   "my-secret",
    Headers:  []string{"date", "host", "@request-target", "digest"},
}
```

Common header options:

| Header | Description |
|--------|-------------|
| `date` | Request time (GMT format), used by Kong for clock skew validation |
| `host` | Request target host |
| `@request-target` | Request method + path + query parameters |
| `digest` | Request body SHA-256 digest for body integrity verification |
| Custom header | e.g., `x-request-id`, `x-custom-header`, etc. |

> When `Headers` includes `digest` and Kong has `validate_request_body` enabled, the SDK automatically computes and injects the `Digest` header. Form / multipart bodies skip Digest computation.

### Supported Algorithms

| Algorithm | Constant | Description |
|-----------|----------|-------------|
| `hmac-sha1` | `AlgorithmHmacSHA1` | SHA-1, disabled by default in Kong, not recommended |
| `hmac-sha224` | `AlgorithmHmacSHA224` | Requires Kong 3.14+ |
| `hmac-sha256` | `AlgorithmHmacSHA256` | **Recommended default** |
| `hmac-sha384` | `AlgorithmHmacSHA384` | |
| `hmac-sha512` | `AlgorithmHmacSHA512` | |

### Choosing Between Do and DoRaw

| Scenario | Recommended Method | Description |
|----------|-------------------|-------------|
| POST JSON body | `Do` | Struct with `json` tag, auto-serialized |
| GET with query params | `Do` | Struct with `form` tag, auto-assembled query string |
| GET with path params | `Do` | Struct with `path` tag, auto-fills `:id` |
| Set custom headers | `Do` | Struct with `header` tag |
| GET without params | `Do` | Pass `nil` for data |
| File upload (multipart) | `DoRaw` | Build multipart body manually |
| XML body | `DoRaw` | Pass XML string as bytes |
| Plain text / encrypted payload | `DoRaw` | Pass raw bytes |
| Form (manually encoded) | `DoRaw` | Pass pre-encoded bytes |

### Do's data Parameter Tag Reference

`Do` delegates to go-zero httpc `buildRequest`, which auto-maps via struct tags:

| Tag | Purpose | Example |
|-----|---------|---------|
| `path:"id"` | Fill URL path parameter `:id` | `/v1/users/:id` → `/v1/users/123` |
| `form:"page"` | Append to query string | `?page=1` |
| `json:"name"` | JSON body field | `{"name":"tom"}` |
| `header:"X-Token"` | Request header | `X-Token: abc` |

> When the struct has a `json` tag, httpc automatically sets `Content-Type: application/json`. GET / HEAD methods must not have a body.

### Timeout Control

The SDK follows go-zero httpc conventions — **no `http.Client.Timeout` is set; timeouts are fully controlled by `context.Context`**.

In the go-zero framework, the `ctx` passed by the caller already has a deadline (injected by `rest.RestConf.Timeout`), which automatically propagates to each request with no extra handling needed:

```go
// In a go-zero handler, l.ctx already has RestConf.Timeout's deadline
resp, err := l.svcCtx.Kong.Do(l.ctx, http.MethodGet, "/v1/users", nil)
```

When used outside go-zero, pass a deadline via `context.WithTimeout`:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

resp, err := client.Do(ctx, http.MethodGet, "/v1/users", nil)
```

> **Note**: Do not set `Timeout` on a custom `*http.Client`. If `http.Client.Timeout` expires before the context deadline, it overrides the caller's intended timeout, causing inconsistent behavior.

## Full Examples

### Create a Client

```go
import "github.com/lerity-yao/czt-contrib/kong/hmacauth"

client := hmacauth.MustNewClient(hmacauth.Conf{
    Host:     "https://api.example.com",
    Username: "alice",
    Secret:   "my-secret",
})
```

### Custom Algorithm and Signing Headers

```go
client := hmacauth.MustNewClient(hmacauth.Conf{
    Host:      "https://api.example.com",
    Username:  "alice",
    Secret:    "my-secret",
    Algorithm: hmacauth.AlgorithmHmacSHA512,
    Headers:   []string{"date", "host", "@request-target", "digest"},
})
```

### Do: Structured Requests

```go
// 1. POST JSON body + parse response
type CreateUserReq struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

resp, err := client.Do(ctx, http.MethodPost, "/v1/users", &CreateUserReq{
    Name:  "tom",
    Email: "tom@example.com",
})

var result struct {
    ID int `json:"id"`
}
if err := hmacauth.Parse(resp, err, &result); err != nil {
    return err
}
// result.ID == 123, resp.Body is automatically closed
```

```go
// 2. GET with path + query params
type GetUserReq struct {
    ID   string `path:"id"`
    Page int    `form:"page"`
}

// Auto-fills: GET /v1/users/123?page=1
resp, err := client.Do(ctx, http.MethodGet, "/v1/users/:id", GetUserReq{
    ID:   "123",
    Page: 1,
})
defer resp.Body.Close()
```

```go
// 3. GET without params, pass nil for data
resp, err := client.Do(ctx, http.MethodGet, "/v1/users", nil)
defer resp.Body.Close()
```

### DoRaw: Raw Requests

```go
// 1. Raw JSON bytes
jsonBody, _ := json.Marshal(map[string]any{"name": "tom"})
resp, err := client.DoRaw(ctx, http.MethodPost, "/v1/users",
    "application/json", jsonBody)
defer resp.Body.Close()
```

```go
// 2. XML body
xmlStr := `<?xml version="1.0"?><user><name>tom</name></user>`
resp, err := client.DoRaw(ctx, http.MethodPost, "/api/pay",
    "application/xml", []byte(xmlStr))
defer resp.Body.Close()
```

```go
// 3. File upload (multipart)
multipartBody := &bytes.Buffer{}
writer := multipart.NewWriter(multipartBody)
part, _ := writer.CreateFormFile("file", "test.png")
part.Write(fileBytes)
writer.Close()

resp, err := client.DoRaw(ctx, http.MethodPost, "/upload",
    writer.FormDataContentType(), multipartBody.Bytes())
defer resp.Body.Close()
```

```go
// 4. GET without body, pass empty contentType
resp, err := client.DoRaw(ctx, http.MethodGet, "/v1/users", "", nil)
defer resp.Body.Close()
```

### Custom HTTP Client

```go
import "net/http"
import "crypto/tls"

cli := &http.Client{
    Transport: &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
        MaxIdleConns:    100,
    },
}

client := hmacauth.MustNewClient(conf, hmacauth.WithClient(cli))
```

### Using with go-zero

```go
// internal/config/config.go
type Config struct {
    rest.RestConf
    Kong hmacauth.Conf
}
```

```yaml
# etc/config.yaml
Name: order-api
Host: 0.0.0.0
Port: 8888

Kong:
  Host: https://api.example.com
  Username: alice
  Secret: my-secret
  # Algorithm and Headers can be omitted, using defaults
  # Algorithm: hmac-sha256
  # Headers:
  #   - date
  #   - host
  #   - "@request-target"
  #   - digest
```

```go
// internal/svc/servicecontext.go
type ServiceContext struct {
    Config config.Config
    Kong   hmacauth.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
    return &ServiceContext{
        Config: c,
        Kong:   hmacauth.MustNewClient(c.Kong),
    }
}
```

```go
// internal/logic/createuserlogic.go
func (l *CreateUserLogic) CreateUser(req *types.CreateUserReq) error {
    resp, err := l.svcCtx.Kong.Do(l.ctx, http.MethodPost, "/v1/users", &struct {
        Name  string `json:"name"`
        Email string `json:"email"`
    }{
        Name:  req.Name,
        Email: req.Email,
    })
    if err != nil {
        return err
    }
    defer resp.Body.Close()

    // Process response...
    return nil
}
```

## Changelog

See [CHANGELOG.md](./CHANGELOG.md)
