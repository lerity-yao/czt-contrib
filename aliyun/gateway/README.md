# gateway

[中文](./readme-cn.md)

A Go client for Alibaba Cloud API Gateway built on [go-zero](https://github.com/zeromicro/go-zero) httpc, with automatic **HMAC-SHA256 v1 signing** and built-in **circuit breaker**.

## Features

- 🔐 **Auto Signing** — Automatically injects `X-Ca-*` signature headers (AppKey, Nonce, Timestamp, Signature) into every request, transparent to the caller
- 🔀 **Dual-method API** — `Do` for structured data (JSON / form / path / header auto-mapping), `DoRaw` for raw bytes (file uploads, XML, plain text, etc.)
- 🛡️ **Circuit Breaker** — Built on go-zero httpc, automatically shares circuit breakers per Host, trips when error rate is too high
- 🔧 **Custom HTTP Client** — Inject a custom `*http.Client` via `WithClient` (timeout, TLS, connection pool, etc.); defaults to `http.DefaultClient` if not provided

## Installation

```bash
go get github.com/lerity-yao/czt-contrib/aliyun/gateway@v0.0.3
```

## Configuration

### Conf

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `Host` | string | Yes | Gateway address, must start with `http://` or `https://` |
| `AppKey` | string | Yes | AppKey of the API Gateway application |
| `AppSecret` | string | Yes | AppSecret of the API Gateway application, used for HMAC-SHA256 signing |

Calling `NewClient` automatically runs `Validate()` to verify the above fields.

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

### Helper Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `NewMultipart` | `NewMultipart() *MultipartBuilder` | Build a multipart/form-data request body |

**MultipartBuilder Methods**

| Method | Description |
|--------|-------------|
| `Field(name, value string)` | Add a text field |
| `File(name, filename string, content []byte)` | Add a file field |
| `Build() (contentType string, body []byte, err error)` | Returns Content-Type, body, and error; can be passed directly to `DoRaw` |

> Use standard library `http.MethodGet`, `http.MethodPost`, etc. for HTTP methods.

## Advanced Guide

### Signing Mechanism

The SDK is based on the [Alibaba Cloud API Gateway V1 signing algorithm](https://help.aliyun.com/document_detail/29475.html). Each request automatically performs the following:

1. **Set default headers** (when not already set): `Accept`, `Date`, `User-Agent`
2. **Calculate Content-MD5**: When body is non-empty and Content-Type is not form / multipart, compute MD5 Base64 of the body
3. **Set X-Ca-* headers**: `X-Ca-Key` (AppKey), `X-Ca-Nonce` (UUID), `X-Ca-Signature-Method` (HmacSHA256), `X-Ca-Timestamp` (millisecond timestamp)
4. **Build the string to sign**:

```
HTTPMethod\n
Accept\n
Content-MD5\n
Content-Type\n
Date\n
X-Ca-headers (alphabetically sorted lowercase key:value)\n
URL (path + alphabetically sorted query)
```

5. **Calculate signature**: HMAC-SHA256 the string to sign with AppSecret, Base64-encode the result, and set it in the `X-Ca-Signature` header

> **Query parameter signing rules**: The query in the signing string uses the original encoded value (`r.URL.RawQuery`), sorted by key alphabetically. For duplicate keys, only the first value is taken; when the value is empty, only the key is kept (without `=`). No decoding or re-encoding is performed, ensuring consistency with the gateway server's verification.

### Choosing Between Do and DoRaw

| Scenario | Recommended Method | Description |
|----------|-------------------|-------------|
| POST JSON body | `Do` | Struct with `json` tag, auto-serialized |
| GET with query params | `Do` | Struct with `form` tag, auto-assembled query string |
| GET with path params | `Do` | Struct with `path` tag, auto-fills `:id` |
| Set custom headers | `Do` | Struct with `header` tag |
| GET without params | `Do` | Pass `nil` for data |
| File upload (multipart) | `DoRaw` | Build with `NewMultipart()`, or manually |
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
resp, err := l.svcCtx.Gateway.Do(l.ctx, http.MethodPost, "/v1/users", data)
```

When used outside go-zero, pass a deadline via `context.WithTimeout`:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

resp, err := client.Do(ctx, http.MethodPost, "/v1/users", data)
```

> **Note**: Do not set `Timeout` on a custom `*http.Client`. If `http.Client.Timeout` expires before the context deadline, it overrides the caller's intended timeout, causing inconsistent behavior.

## Full Examples

### Create a Client

```go
import "github.com/lerity-yao/czt-contrib/aliyun/gateway"

client := gateway.MustNewClient(gateway.Conf{
    Host:      "https://your-gateway.aliyuncs.com",
    AppKey:    "your-app-key",
    AppSecret: "your-app-secret",
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
if err := gateway.Parse(resp, err, &result); err != nil {
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
// 3. File upload (multipart) — using MultipartBuilder
ct, body, err := gateway.NewMultipart().
    Field("description", "avatar").
    File("file", "test.png", fileBytes).
    Build()
if err != nil {
    return err
}

resp, err := client.DoRaw(ctx, http.MethodPost, "/upload", ct, body)
defer resp.Body.Close()
```

```go
// 3b. File upload (multipart) — native approach
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

client := gateway.MustNewClient(conf, gateway.WithClient(cli))
```

### Using with go-zero

```go
// internal/config/config.go
type Config struct {
    rest.RestConf
    Gateway gateway.Conf
}
```

```yaml
# etc/config.yaml
Name: order-api
Host: 0.0.0.0
Port: 8888

Gateway:
  Host: https://your-gateway.aliyuncs.com
  AppKey: your-app-key
  AppSecret: your-app-secret
```

```go
// internal/svc/servicecontext.go
type ServiceContext struct {
    Config    config.Config
    Gateway   gateway.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
    return &ServiceContext{
        Config:  c,
        Gateway: gateway.MustNewClient(c.Gateway),
    }
}
```

```go
// internal/logic/createuserlogic.go
func (l *CreateUserLogic) CreateUser(req *types.CreateUserReq) error {
    resp, err := l.svcCtx.Gateway.Do(l.ctx, http.MethodPost, "/v1/users", &struct {
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
