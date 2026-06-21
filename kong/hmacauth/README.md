# hmacauth

English | [中文](./readme-cn.md)

A Kong HMAC Auth Go client built on top of [go-zero](https://github.com/zeromicro/go-zero) httpc, providing automatic **HMAC signing** and a built-in **circuit breaker**. Complies with the [Kong HMAC Auth Plugin](https://developer.konghq.com/plugins/hmac-auth/) official specification (based on the [draft-cavage-http-signatures](https://tools.ietf.org/html/draft-cavage-http-signatures) draft, using the `@request-target` pseudo-header).

## Features

- 🔐 **Automatic signing** — Automatically injects the `Authorization` signature header (`hmac username="...", algorithm="...", headers="...", signature="..."`) into every request, completely transparent to the caller
- 🔀 **Dual-method API** — `Do` handles structured data (automatic JSON / form / path / header mapping), `DoRaw` handles raw bytes (file uploads, XML, plain text, etc.)
- 🛡️ **Circuit breaker protection** — Powered by go-zero httpc; requests to the same host automatically share a circuit breaker that trips when the error rate is too high
- 🔧 **Customizable HTTP client** — Inject a custom `*http.Client` (timeout, TLS, connection pool, etc.) via `WithClient`; falls back to `http.DefaultClient` if not provided
- 🧩 **Multi-algorithm support** — Full coverage of five algorithms: `hmac-sha1`, `hmac-sha224`, `hmac-sha256`, `hmac-sha384`, `hmac-sha512`
- 📋 **Flexible signing headers** — Supports a customizable list of headers to include in the signature; defaults to `["date", "@request-target"]`
- 📦 **Body integrity** — When `Headers` includes `digest`, automatically computes and injects the `Digest: SHA-256=...` header (including empty-body scenarios)

## Installation

```bash
go get github.com/lerity-yao/czt-contrib/kong/hmacauth@v0.0.1
```

## Configuration

### Conf

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `Host` | string | Yes | — | Kong gateway address; must start with `http://` or `https://` |
| `Username` | string | Yes | — | Kong consumer credential username (i.e., key ID) |
| `Secret` | string | Yes | — | HMAC signing secret |
| `Algorithm` | string | No | `hmac-sha256` | Signing algorithm; supports `hmac-sha1` / `hmac-sha224` / `hmac-sha256` / `hmac-sha384` / `hmac-sha512` |
| `Headers` | []string | No | `["date", "@request-target"]` | List of headers (lowercase) to include in the signature; supports standard HTTP headers and the `@request-target` pseudo-header |

`Validate()` is called automatically when `NewClient` is invoked, validating the fields above and filling in default values for `Algorithm` and `Headers`.

> `Algorithm` and `Headers` support go-zero `conf.MustLoad`'s `optional` / `default` tags, so they can be omitted from YAML configuration files.

## API Reference

### Constructors

| Function | Signature | Description |
|----------|-----------|-------------|
| `MustNewClient` | `func MustNewClient(c Conf, opts ...ClientOption) Client` | Creates a Client; panics on validation failure |
| `NewClient` | `func NewClient(c Conf, opts ...ClientOption) (Client, error)` | Creates a Client; returns an error on validation failure |

### ClientOption

| Option | Parameter | Description |
|--------|-----------|-------------|
| `WithClient` | `*http.Client` | Injects a custom HTTP client (timeout, TLS, connection pool, etc.); falls back to `http.DefaultClient` if not provided |

### Algorithm Constants

| Constant | Value | Description |
|----------|-------|-------------|
| `AlgorithmHmacSHA1` | `"hmac-sha1"` | SHA-1; disabled by default in Kong — not recommended |
| `AlgorithmHmacSHA224` | `"hmac-sha224"` | Requires Kong 3.14+ |
| `AlgorithmHmacSHA256` | `"hmac-sha256"` | **Recommended default** |
| `AlgorithmHmacSHA384` | `"hmac-sha384"` | |
| `AlgorithmHmacSHA512` | `"hmac-sha512"` | |

### Client Interface Methods

| Method | Signature | Use Case |
|--------|-----------|----------|
| `Do` | `Do(ctx, method, path string, data any) (*http.Response, error)` | Structured requests: automatic JSON body / form query / path parameter / header mapping |
| `DoRaw` | `DoRaw(ctx, method, path, contentType string, body []byte) (*http.Response, error)` | Raw requests: file uploads, XML, plain text, encrypted payloads, and other custom bodies |

> **Note**: `DoRaw` **requires** a non-empty `contentType` when `body` is non-nil; otherwise it returns an error. For requests without a body (e.g., GET), `contentType` may be empty.

### Response Parsing

| Function | Signature | Use Case |
|----------|-----------|----------|
| `Parse` | `Parse(resp *http.Response, err error, val any) error` | JSON responses: parses headers + body and automatically closes Body |

`Parse` delegates internally to go-zero `httpc.Parse`, supporting response header (`header` tag) and JSON body (`json` tag) parsing. For non-JSON responses such as XML or binary data, read `resp.Body` directly.

> For HTTP methods, use the standard library constants `http.MethodGet`, `http.MethodPost`, etc.

## Advanced Guide

### Signing Mechanism

The SDK follows the [Kong HMAC Auth Plugin](https://developer.konghq.com/plugins/hmac-auth/) official specification and automatically performs the following steps for each request:

1. **Set default headers** (if not already set): `Date` (GMT format), `User-Agent` (defaults to `Go-Kong-HmacAuth-Client`), `Host` (when `host` is included in the signing headers list)
2. **Compute Digest** (when `Headers` includes `digest`): computes `SHA-256` over non-form / non-multipart bodies, formatted as `Digest: SHA-256=<base64>`; empty bodies are also digested
3. **Build the signing string**: concatenates headers in the order specified by `Headers`, one per line in `header-name: value` format, separated by `\n` with no trailing newline

```
date: Thu, 22 Jun 2023 17:15:21 GMT
@request-target: get /v1/users?page=1
host: api.example.com
digest: SHA-256=47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=
```

4. **Compute the signature**: performs HMAC over the signing string using `Secret` (with the algorithm specified by `Algorithm`), then Base64-encodes the result
5. **Set the Authorization header**:

```
Authorization: hmac username="alice", algorithm="hmac-sha256", headers="date @request-target", signature="abc123..."
```

### The @request-target Pseudo-Header

Kong HMAC Auth uses `@request-target` as a pseudo-header representing the request target line. Its value is `method /path?query` (method in lowercase):

```
@request-target: post /v1/users?name=tom
```

> Kong officially recommends using `@request-target` rather than the legacy `request-line`.

### Digest Computation

When `Headers` includes `digest`, the SDK automatically computes and injects the `Digest` header:

- **Algorithm**: always uses SHA-256, formatted as `Digest: SHA-256=<base64>`
- **Empty body**: the digest is still computed (`sha256.Sum256(nil)`), ensuring request integrity is verifiable
- **form / multipart skipped**: bodies with `application/x-www-form-urlencoded` or `multipart/form-data` content types are excluded from Digest computation, because form data may be re-encoded by the HTTP transport before it reaches the server

> When Kong's `validate_request_body` option is also enabled, the SDK's Digest header ensures end-to-end body integrity verification.

### Algorithm Selection

| Algorithm | Security | Performance | Use Case |
|-----------|----------|-------------|----------|
| `hmac-sha1` | Weak (deprecated) | Fast | Legacy system compatibility only; disabled by default in Kong |
| `hmac-sha224` | Moderate | Faster | Requires Kong 3.14+ |
| `hmac-sha256` | Strong | Fast | **Recommended default** — best balance of security and performance |
| `hmac-sha384` | Very strong | Slower | High-security scenarios |
| `hmac-sha512` | Very strong | Slowest | Maximum-security scenarios |

> The algorithm is specified via `Conf.Algorithm` or a constant (e.g., `hmacauth.AlgorithmHmacSHA256`); defaults to `hmac-sha256` if not set. The value is case-insensitive and is normalized to lowercase internally.

### Signing Headers Configuration

`Headers` determines which headers participate in the signature. By default, only `date` and `@request-target` are signed. The strongly recommended Kong signing configuration is `@request-target` + `host` + `date` + `digest`:

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
| `date` | Request timestamp (GMT format); Kong uses this for clock skew validation |
| `host` | Target host of the request |
| `@request-target` | Request method + path + query parameters |
| `digest` | SHA-256 digest of the request body; used for body integrity verification |
| Custom headers | e.g., `x-request-id`, `x-custom-header`, etc. |

> Header names in `Headers` are automatically normalized to lowercase. For custom headers, ensure the corresponding value is set on the request; otherwise the header value in the signing string will be empty.

### Choosing Between Do and DoRaw

| Scenario | Recommended Method | Description |
|----------|--------------------|-------------|
| POST JSON body | `Do` | Struct with `json` tags; automatically serialized |
| GET with query parameters | `Do` | Struct with `form` tags; automatically appended to query string |
| GET with path parameters | `Do` | Struct with `path` tags; automatically fills `:id` placeholders |
| Set custom headers | `Do` | Struct with `header` tags |
| GET with no parameters | `Do` | Pass `nil` for `data` |
| File upload (multipart) | `DoRaw` | Manually construct the multipart body |
| XML body | `DoRaw` | Pass the XML string as bytes |
| Plain text / encrypted payload | `DoRaw` | Pass raw bytes |
| Form (manually encoded) | `DoRaw` | Pass pre-encoded bytes |

### Struct Tag Reference for Do's data Parameter

`Do` delegates internally to go-zero httpc `buildRequest`, which maps struct tags automatically:

| Tag | Effect | Example |
|-----|--------|---------|
| `path:"id"` | Fills a URL path parameter `:id` | `/v1/users/:id` → `/v1/users/123` |
| `form:"page"` | Appended to query string | `?page=1` |
| `json:"name"` | JSON body field | `{"name":"tom"}` |
| `header:"X-Token"` | Request header | `X-Token: abc` |

> When the struct contains `json` tags, httpc automatically sets `Content-Type: application/json`. GET / HEAD methods do not allow a request body.

### Timeout Control

The SDK follows the go-zero httpc convention: **`http.Client.Timeout` is not set; timeouts are controlled entirely by `context.Context`**.

In a go-zero service, the `ctx` passed by the caller already carries a deadline (injected by `rest.RestConf.Timeout`), which propagates automatically to every request with no additional configuration needed:

```go
// In a go-zero handler, l.ctx already carries the RestConf.Timeout deadline
resp, err := l.svcCtx.Kong.Do(l.ctx, http.MethodGet, "/v1/users", nil)
```

When using the SDK outside of a go-zero service, pass a deadline via `context.WithTimeout`:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

resp, err := client.Do(ctx, http.MethodGet, "/v1/users", nil)
```

> **Note**: Do not set `Timeout` on a custom `*http.Client`. If `http.Client.Timeout` expires before the context deadline, it overrides the caller's intended timeout, leading to inconsistent behavior.

## Full Examples

### Creating a Client

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
// result.ID == 123; resp.Body is closed automatically
```

```go
// 2. GET with path + query parameters
type GetUserReq struct {
    ID   string `path:"id"`
    Page int    `form:"page"`
}

// Automatically resolves to: GET /v1/users/123?page=1
resp, err := client.Do(ctx, http.MethodGet, "/v1/users/:id", GetUserReq{
    ID:   "123",
    Page: 1,
})
defer resp.Body.Close()
```

```go
// 3. GET with no parameters; pass nil for data
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
// 4. GET with no body; pass empty string for contentType
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

### Usage in go-zero

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
  # Algorithm and Headers can be omitted to use default values
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

    // Handle response...
    return nil
}
```

### Usage in a Standalone Script

```go
package main

import (
    "context"
    "fmt"
    "net/http"
    "time"

    "github.com/lerity-yao/czt-contrib/kong/hmacauth"
)

func main() {
    client := hmacauth.MustNewClient(hmacauth.Conf{
        Host:      "https://api.example.com",
        Username:  "alice",
        Secret:    "my-secret",
        Algorithm: hmacauth.AlgorithmHmacSHA256,
    })

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // GET request
    resp, err := client.Do(ctx, http.MethodGet, "/v1/users", nil)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    fmt.Println("Status:", resp.StatusCode)

    // POST JSON request
    resp, err = client.Do(ctx, http.MethodPost, "/v1/users", &struct {
        Name string `json:"name"`
    }{Name: "tom"})
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    var result struct {
        ID int `json:"id"`
    }
    if err := hmacauth.Parse(resp, err, &result); err != nil {
        panic(err)
    }
    fmt.Println("Created user ID:", result.ID)
}
```

## Changelog

See [CHANGELOG.md](./CHANGELOG.md)
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
