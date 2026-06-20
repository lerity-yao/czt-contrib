# hmacauth

[English](./README.md)

基于 [go-zero](https://github.com/zeromicro/go-zero) httpc 封装的 Kong HMAC Auth Go 客户端，自动完成 **HMAC 签名**，内置**熔断器**。

遵循 [Kong HMAC Auth 插件](https://developer.konghq.com/plugins/hmac-auth/) 官方规范（基于 [draft-cavage-http-signatures](https://tools.ietf.org/html/draft-cavage-http-signatures) 草案，使用 `@request-target` 伪头）。

## 特性

- 🔐 **自动签名** — 每个请求自动注入 `Authorization` 签名头（`hmac username="...", algorithm="...", headers="...", signature="..."`），调用方无感知
- 🔀 **双方法 API** — `Do` 处理结构化数据（JSON / form / path / header 自动映射），`DoRaw` 处理原始字节（文件上传、XML、纯文本等）
- 🛡️ **熔断保护** — 底层使用 go-zero httpc，同一个 Host 自动共享熔断器，错误率过高自动熔断
- 🔧 **可定制 HTTP Client** — 通过 `WithClient` 注入自定义 `*http.Client`（超时、TLS、连接池等），不传则使用 `http.DefaultClient`
- 🧩 **多算法支持** — `hmac-sha1`、`hmac-sha224`、`hmac-sha256`、`hmac-sha384`、`hmac-sha512` 五种算法全覆盖
- 📋 **灵活签名头** — 支持自定义参与签名的 header 列表，默认 `["date", "@request-target"]`
- 📦 **Body 完整性** — 当 `Headers` 包含 `digest` 时，自动计算 `Digest: SHA-256=...` 头（含空 body 场景）

## 安装

```bash
go get github.com/lerity-yao/czt-contrib/kong/hmacauth@v0.0.1
```

## 配置参数

### Conf

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| `Host` | string | 是 | — | Kong 网关地址，必须以 `http://` 或 `https://` 开头 |
| `Username` | string | 是 | — | Kong consumer 凭证用户名（即 key id） |
| `Secret` | string | 是 | — | HMAC 签名密钥 |
| `Algorithm` | string | 否 | `hmac-sha256` | 签名算法，支持 `hmac-sha1` / `hmac-sha224` / `hmac-sha256` / `hmac-sha384` / `hmac-sha512` |
| `Headers` | []string | 否 | `["date", "@request-target"]` | 参与签名的 header 列表（小写），支持标准 HTTP header 和 `@request-target` 伪头 |

调用 `NewClient` 时会自动执行 `Validate()` 校验上述字段，并为 `Algorithm` 和 `Headers` 填充默认值。

> `Algorithm` 和 `Headers` 支持 go-zero `conf.MustLoad` 的 `optional` / `default` tag，YAML 配置文件中可省略。

## API 参考

### 构造函数

| 函数 | 签名 | 说明 |
|------|------|------|
| `MustNewClient` | `func MustNewClient(c Conf, opts ...ClientOption) Client` | 创建 Client，校验失败 panic |
| `NewClient` | `func NewClient(c Conf, opts ...ClientOption) (Client, error)` | 创建 Client，校验失败返回 error |

### ClientOption

| Option | 参数 | 说明 |
|--------|------|------|
| `WithClient` | `*http.Client` | 注入自定义 HTTP Client（超时、TLS、连接池等），不传则使用 `http.DefaultClient` |

### Client 接口方法

| 方法 | 签名 | 适用场景 |
|------|------|----------|
| `Do` | `Do(ctx, method, path string, data any) (*http.Response, error)` | 结构化请求：JSON body / form query / path 参数 / header 自动映射 |
| `DoRaw` | `DoRaw(ctx, method, path, contentType string, body []byte) (*http.Response, error)` | 原始请求：文件上传、XML、纯文本、加密 payload 等自定义 body |

> **注意**：`DoRaw` 在 `body` 非空时**强制要求** `contentType` 非空，否则返回 error。无 body 的请求（如 GET）`contentType` 可为空。

### 响应解析

| 函数 | 签名 | 适用场景 |
|------|------|----------|
| `Parse` | `Parse(resp *http.Response, err error, val any) error` | JSON 响应：解析 header + body，自动关闭 Body |

`Parse` 内部委托 go-zero `httpc.Parse`，支持响应头（`header` tag）和 JSON body（`json` tag）解析。对于 XML、二进制等非 JSON 响应，请直接读取 `resp.Body`。

> HTTP 方法请直接使用标准库 `http.MethodGet`、`http.MethodPost` 等。

## 进阶指南

### 签名机制

SDK 基于 [Kong HMAC Auth 插件](https://developer.konghq.com/plugins/hmac-auth/) 官方规范，每个请求自动完成以下工作：

1. **设置默认头**（未设置时）：`Date`（GMT 格式）、`User-Agent`、`Host`（当 host 在签名列表中时）
2. **计算 Digest**（当 `Headers` 包含 `digest` 时）：对非 form / multipart 的 body 计算 `SHA-256`，格式为 `Digest: SHA-256=<base64>`；空 body 也计算零长度摘要
3. **构建待签名字符串**：按 `Headers` 列表顺序拼接，每行格式为 `header-name: value`，行间用 `\n` 分隔，最后一行无尾随换行

```
date: Thu, 22 Jun 2023 17:15:21 GMT
@request-target: get /v1/users?page=1
host: api.example.com
digest: SHA-256=47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=
```

4. **计算签名**：用 `Secret` 对待签名字符串做 HMAC（算法由 `Algorithm` 指定），结果 Base64 编码
5. **设置 Authorization 头**：

```
Authorization: hmac username="alice", algorithm="hmac-sha256", headers="date @request-target", signature="abc123..."
```

### @request-target 伪头

Kong HMAC Auth 使用 `@request-target` 作为伪头表示请求目标行，值为 `method /path?query`（method 全小写）：

```
@request-target: post /v1/users?name=tom
```

> Kong 官方推荐使用 `@request-target` 而非旧版的 `request-line`。

### 签名头列表配置

`Headers` 决定哪些 header 参与签名。默认只签 `date` 和 `@request-target`，Kong 官方推荐的强签名配置为 `@request-target` + `host` + `date`：

```go
hmacauth.Conf{
    Host:     "https://api.example.com",
    Username: "alice",
    Secret:   "my-secret",
    Headers:  []string{"date", "host", "@request-target", "digest"},
}
```

常见 header 选项：

| Header | 说明 |
|--------|------|
| `date` | 请求时间（GMT 格式），Kong 用此做 clock skew 校验 |
| `host` | 请求目标主机 |
| `@request-target` | 请求方法 + 路径 + 查询参数 |
| `digest` | 请求体 SHA-256 摘要，用于 body 完整性校验 |
| 自定义 header | 如 `x-request-id`、`x-custom-header` 等 |

> 当 `Headers` 包含 `digest` 且 Kong 端启用了 `validate_request_body` 时，SDK 会自动计算并注入 `Digest` 头。form / multipart body 跳过 Digest 计算。

### 支持的算法

| 算法 | 常量 | 说明 |
|------|------|------|
| `hmac-sha1` | `AlgorithmHmacSHA1` | SHA-1，Kong 默认禁用，不推荐使用 |
| `hmac-sha224` | `AlgorithmHmacSHA224` | 需要 Kong 3.14+ |
| `hmac-sha256` | `AlgorithmHmacSHA256` | **推荐默认** |
| `hmac-sha384` | `AlgorithmHmacSHA384` | |
| `hmac-sha512` | `AlgorithmHmacSHA512` | |

### Do 与 DoRaw 如何选择

| 场景 | 推荐方法 | 说明 |
|------|----------|------|
| POST JSON body | `Do` | struct 带 `json` tag，自动序列化 |
| GET 带 query 参数 | `Do` | struct 带 `form` tag，自动拼 query string |
| GET 带 path 参数 | `Do` | struct 带 `path` tag，自动填充 `:id` |
| 设置自定义 header | `Do` | struct 带 `header` tag |
| GET 无参数 | `Do` | data 传 `nil` |
| 文件上传（multipart） | `DoRaw` | 手动构建 multipart body |
| XML body | `DoRaw` | 传 XML 字符串的字节 |
| 纯文本 / 加密 payload | `DoRaw` | 传原始字节 |
| 表单（手动编码） | `DoRaw` | 传已编码的字节 |

### Do 的 data 参数 Tag 说明

`Do` 内部委托 go-zero httpc `buildRequest`，通过 struct tag 自动映射：

| Tag | 作用 | 示例 |
|-----|------|------|
| `path:"id"` | 填充 URL 路径参数 `:id` | `/v1/users/:id` → `/v1/users/123` |
| `form:"page"` | 拼接到 query string | `?page=1` |
| `json:"name"` | JSON body 字段 | `{"name":"tom"}` |
| `header:"X-Token"` | 请求头 | `X-Token: abc` |

> 当 struct 含 `json` tag 时，httpc 自动设置 `Content-Type: application/json`。GET / HEAD 方法不允许带 body。

### 超时控制

SDK 遵循 go-zero httpc 惯例，**不设 `http.Client.Timeout`，超时完全由 `context.Context` 控制**。

在 go-zero 框架中，调用方传入的 `ctx` 已自带 deadline（由 `rest.RestConf.Timeout` 注入），会自动传播到每个请求，无需额外处理：

```go
// go-zero handler 中，l.ctx 已有 RestConf.Timeout 的 deadline
resp, err := l.svcCtx.Kong.Do(l.ctx, http.MethodGet, "/v1/users", nil)
```

如果在非 go-zero 环境中使用，请通过 `context.WithTimeout` 传入 deadline：

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

resp, err := client.Do(ctx, http.MethodGet, "/v1/users", nil)
```

> **注意**：不要在自定义 `*http.Client` 上设置 `Timeout`。如果 `http.Client.Timeout` 早于 context deadline 到期，会覆盖调用方意图中的超时时间，导致行为与预期不一致。

## 完整示例

### 创建客户端

```go
import "github.com/lerity-yao/czt-contrib/kong/hmacauth"

client := hmacauth.MustNewClient(hmacauth.Conf{
    Host:     "https://api.example.com",
    Username: "alice",
    Secret:   "my-secret",
})
```

### 自定义算法和签名头

```go
client := hmacauth.MustNewClient(hmacauth.Conf{
    Host:      "https://api.example.com",
    Username:  "alice",
    Secret:    "my-secret",
    Algorithm: hmacauth.AlgorithmHmacSHA512,
    Headers:   []string{"date", "host", "@request-target", "digest"},
})
```

### Do：结构化请求

```go
// 1. POST JSON body + 解析响应
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
// result.ID == 123，resp.Body 已自动关闭
```

```go
// 2. GET 带 path + query 参数
type GetUserReq struct {
    ID   string `path:"id"`
    Page int    `form:"page"`
}

// 自动填充：GET /v1/users/123?page=1
resp, err := client.Do(ctx, http.MethodGet, "/v1/users/:id", GetUserReq{
    ID:   "123",
    Page: 1,
})
defer resp.Body.Close()
```

```go
// 3. GET 无参数，data 传 nil
resp, err := client.Do(ctx, http.MethodGet, "/v1/users", nil)
defer resp.Body.Close()
```

### DoRaw：原始请求

```go
// 1. JSON 原始字节
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
// 3. 文件上传（multipart）
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
// 4. GET 无 body，contentType 传空
resp, err := client.DoRaw(ctx, http.MethodGet, "/v1/users", "", nil)
defer resp.Body.Close()
```

### 自定义 HTTP Client

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

### 在 go-zero 中使用

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
  # Algorithm 和 Headers 可省略，使用默认值
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

    // 处理响应...
    return nil
}
```

## 更新日志

查看 [CHANGELOG.md](./CHANGELOG.md)
