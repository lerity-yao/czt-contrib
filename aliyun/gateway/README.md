# gateway

基于 [go-zero](https://github.com/zeromicro/go-zero) httpc 封装的阿里云 API 网关 Go 客户端，自动完成 **HMAC-SHA256 v1 签名**，内置**熔断器**。

## 特性

- 🔐 **自动签名** — 每个请求自动注入 `X-Ca-*` 签名头（AppKey、Nonce、Timestamp、Signature），调用方无感知
- 🔀 **双方法 API** — `Do` 处理结构化数据（JSON / form / path / header 自动映射），`DoRaw` 处理原始字节（文件上传、XML、纯文本等）
- 🛡️ **熔断保护** — 底层使用 go-zero httpc，同一个 Host 自动共享熔断器，错误率过高自动熔断
- 🔧 **可定制 HTTP Client** — 通过 `WithClient` 注入自定义 `*http.Client`（超时、TLS、连接池等），不传则使用 `http.DefaultClient`

## 安装

```bash
go get github.com/lerity-yao/czt-contrib/aliyun/gateway
```

## 配置参数

### Conf

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `Host` | string | 是 | 网关地址，必须以 `http://` 或 `https://` 开头 |
| `AppKey` | string | 是 | API 网关应用的 AppKey |
| `AppSecret` | string | 是 | API 网关应用的 AppSecret，用于 HMAC-SHA256 签名 |

调用 `NewClient` 时会自动执行 `Validate()` 校验上述字段。

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

### 辅助函数

| 函数 | 签名 | 说明 |
|------|------|------|
| `NewMultipart` | `NewMultipart() *MultipartBuilder` | 构造 multipart/form-data 请求体 |

**MultipartBuilder 方法**

| 方法 | 说明 |
|------|------|
| `Field(name, value string)` | 添加文本字段 |
| `File(name, filename string, content []byte)` | 添加文件字段 |
| `Build() (contentType string, body []byte)` | 返回 Content-Type 和 body，可直接传入 `DoRaw` |

> HTTP 方法请直接使用标准库 `http.MethodGet`、`http.MethodPost` 等。

## 进阶指南

### 签名机制

SDK 基于 [阿里云 API 网关 V1 签名算法](https://help.aliyun.com/document_detail/29475.html)，每个请求自动完成以下工作：

1. **设置默认头**（未设置时）：`Accept`、`Date`、`User-Agent`
2. **计算 Content-MD5**：当 body 非空且 Content-Type 不是 form / multipart 时，计算 body 的 MD5 Base64 值
3. **设置 X-Ca-* 头**：`X-Ca-Key`（AppKey）、`X-Ca-Nonce`（UUID）、`X-Ca-Signature-Method`（HmacSHA256）、`X-Ca-Timestamp`（毫秒时间戳）
4. **构建待签名字符串**：

```
HTTPMethod\n
Accept\n
Content-MD5\n
Content-Type\n
Date\n
X-Ca-headers (按字典序排列的小写 key:value)\n
URL (path + 按字典序排序的 query)
```

5. **计算签名**：用 AppSecret 对待签名字符串做 HMAC-SHA256，结果 Base64 编码后放入 `X-Ca-Signature` 头

> **query 参数签名规则**：签名字符串中的 query 使用原始编码值（`r.URL.RawQuery`），按 key 字典序排序。相同 key 的参数仅取第一个 value；value 为空时仅保留 key（不含 `=`）。不解码也不重新编码，确保与网关服务端验证结果一致。

### Do 与 DoRaw 如何选择

| 场景 | 推荐方法 | 说明 |
|------|----------|------|
| POST JSON body | `Do` | struct 带 `json` tag，自动序列化 |
| GET 带 query 参数 | `Do` | struct 带 `form` tag，自动拼 query string |
| GET 带 path 参数 | `Do` | struct 带 `path` tag，自动填充 `:id` |
| 设置自定义 header | `Do` | struct 带 `header` tag |
| GET 无参数 | `Do` | data 传 `nil` |
| 文件上传（multipart） | `DoRaw` | 使用 `NewMultipart()` 构造，或手动构建 |
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
resp, err := l.svcCtx.Gateway.Do(l.ctx, http.MethodPost, "/v1/users", data)
```

如果在非 go-zero 环境中使用，请通过 `context.WithTimeout` 传入 deadline：

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

resp, err := client.Do(ctx, http.MethodPost, "/v1/users", data)
```

> **注意**：不要在自定义 `*http.Client` 上设置 `Timeout`。如果 `http.Client.Timeout` 早于 context deadline 到期，会覆盖调用方意图中的超时时间，导致行为与预期不一致。

## 完整示例

### 创建客户端

```go
import "github.com/lerity-yao/czt-contrib/aliyun/gateway"

client := gateway.MustNewClient(gateway.Conf{
    Host:      "https://your-gateway.aliyuncs.com",
    AppKey:    "your-app-key",
    AppSecret: "your-app-secret",
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
if err := gateway.Parse(resp, err, &result); err != nil {
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
// 3. 文件上传（multipart）— 使用 MultipartBuilder
ct, body := gateway.NewMultipart().
    Field("description", "avatar").
    File("file", "test.png", fileBytes).
    Build()

resp, err := client.DoRaw(ctx, http.MethodPost, "/upload", ct, body)
defer resp.Body.Close()
```

```go
// 3b. 文件上传（multipart）— 原生写法
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

client := gateway.MustNewClient(conf, gateway.WithClient(cli))
```

### 在 go-zero 中使用

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

    // 处理响应...
    return nil
}
```

## 更新日志

查看 [CHANGELOG.md](./CHANGELOG.md)
