# oss

基于 [go-zero](https://github.com/zeromicro/go-zero) httpc 封装的阿里云 OSS Go 客户端，自动完成 **HMAC-SHA1 V1 签名**，内置**熔断器**。

## 特性

- 🔐 **自动签名** — 每个请求自动注入 `Authorization` 签名头（OSS V1 HMAC-SHA1），调用方无感知
- 📦 **语义化 API** — `PutObject` / `GetObject` / `DeleteObject` / `HeadObject` / `CopyObject` / `ListObjects`，开箱即用
- 🛡️ **熔断保护** — 底层使用 go-zero httpc，同一个 Host 自动共享熔断器，错误率过高自动熔断
- 🔧 **可定制 HTTP Client** — 通过 `WithClient` 注入自定义 `*http.Client`（超时、TLS、连接池等），不传则使用 `http.DefaultClient`

## 安装

```bash
go get github.com/lerity-yao/czt-contrib/aliyun/oss
```

## 配置参数

### Conf

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `Endpoint` | string | 是 | OSS 区域节点，如 `oss-cn-hangzhou.aliyuncs.com`。可加 `http://` 前缀用于本地 MinIO |
| `Bucket` | string | 是 | 存储空间名称 |
| `AccessKeyId` | string | 是 | 阿里云 AccessKey ID |
| `AccessKeySecret` | string | 是 | 阿里云 AccessKey Secret，用于 HMAC-SHA1 签名 |

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
| `PutObject` | `PutObject(ctx, key string, body []byte, opts ...Option) error` | 上传对象 |
| `GetObject` | `GetObject(ctx, key string, opts ...Option) (io.ReadCloser, error)` | 下载对象，返回流式 Reader，调用方负责关闭 |
| `DeleteObject` | `DeleteObject(ctx, key string) error` | 删除对象 |
| `HeadObject` | `HeadObject(ctx, key string) (*ObjectMeta, error)` | 获取对象元信息（大小、类型、ETag、自定义元数据） |
| `CopyObject` | `CopyObject(ctx, destKey, srcKey string) error` | 同 Bucket 内拷贝对象 |
| `ListObjects` | `ListObjects(ctx, opts ...ListOption) (*ListBucketResult, error)` | 列举对象，支持前缀 / 分页 / 分隔符 |
| `Do` | `Do(ctx, method, key string, headers map[string]string, body []byte) (*http.Response, error)` | 原始请求，自动签名，调用方负责关闭 Body |

> **Do 方法**：当 OSS 高层方法不满足需求时（如设置对象 ACL、获取 Bucket 信息等），可通过 `Do` 发送任意签名请求。`key` 为对象路径，传空字符串表示 Bucket 级操作。

### 响应解析

| 函数 | 签名 | 适用场景 |
|------|------|----------|
| `Parse` | `Parse(resp *http.Response, err error, val any) error` | XML 响应：解码 body 到 struct，自动关闭 Body |

`Parse` 使用 `encoding/xml` 解码响应体。对于二进制响应（如 `GetObject` 返回的文件流），请直接读取 `io.ReadCloser`，不要使用 `Parse`。

### Option

用于 `PutObject` / `GetObject` 的请求头定制：

| Option | 参数 | 说明 |
|--------|------|------|
| `WithContentType` | `ct string` | 设置 `Content-Type` 请求头 |
| `WithMeta` | `key, value string` | 设置自定义元数据（自动加 `x-oss-meta-` 前缀） |
| `WithHeader` | `key, value string` | 设置任意请求头 |

### ListOption

用于 `ListObjects` 的查询参数定制：

| Option | 参数 | 说明 |
|--------|------|------|
| `WithPrefix` | `prefix string` | 限定返回对象的 Key 前缀 |
| `WithMarker` | `marker string` | 分页起始 Key（从 `NextMarker` 获取） |
| `WithMaxKeys` | `n int` | 最大返回数量（1–1000，不传默认 100） |
| `WithDelimiter` | `d string` | 分隔符，用于模拟目录结构 |
| `WithEncodingType` | `t string` | Key 编码类型（如 `"url"`） |

> HTTP 方法请直接使用标准库 `http.MethodGet`、`http.MethodPut` 等。

## 进阶指南

### 签名机制

SDK 基于 [阿里云 OSS V1 签名算法](https://help.aliyun.com/document_detail/31957.html)，每个请求自动完成以下工作：

1. **设置默认头**（未设置时）：`Date`（GMT 格式）、`User-Agent`
2. **构建 CanonicalizedOSSHeaders**：收集所有 `x-oss-*` 请求头，按小写 key 字典序排列
3. **构建 CanonicalizedResource**：`/{bucket}/{object-key}` + OSS 子资源（如 `acl`、`uploadId` 等，普通 query 参数不参与签名）
4. **构建待签名字符串**：

```
HTTPMethod\n
Content-MD5\n
Content-Type\n
Date\n
CanonicalizedOSSHeaders (按字典序排列的小写 key:value\n)
CanonicalizedResource
```

5. **计算签名**：用 AccessKeySecret 对待签名字符串做 HMAC-SHA1，结果 Base64 编码后放入 `Authorization` 头：`OSS {AccessKeyId}:{Signature}`

> **子资源签名规则**：只有 OSS 定义的子资源（如 `acl`、`uploads`、`response-content-type` 等）参与签名，普通查询参数（如 `prefix`、`marker`、`max-keys`）不参与。

### 错误处理

当 OSS 返回 HTTP 4xx / 5xx 时，高层方法自动解析 XML 错误体为 `*ServiceError`：

```go
err := client.PutObject(ctx, "key", body)
if err != nil {
    var svcErr *oss.ServiceError
    if errors.As(err, &svcErr) {
        // svcErr.Code: "SignatureDoesNotMatch"
        // svcErr.Message: "The request signature we calculated does not match..."
        // svcErr.RequestId: "xxxx"
    }
}
```

### 超时控制

SDK 遵循 go-zero httpc 惯例，**不设 `http.Client.Timeout`，超时完全由 `context.Context` 控制**。

在 go-zero 框架中，调用方传入的 `ctx` 已自带 deadline（由 `rest.RestConf.Timeout` 注入），会自动传播到每个请求，无需额外处理：

```go
// go-zero handler 中，l.ctx 已有 RestConf.Timeout 的 deadline
rc, err := l.svcCtx.OSS.GetObject(l.ctx, "photos/test.jpg")
```

如果在非 go-zero 环境中使用，请通过 `context.WithTimeout` 传入 deadline：

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

rc, err := client.GetObject(ctx, "photos/test.jpg")
```

> **注意**：不要在自定义 `*http.Client` 上设置 `Timeout`。如果 `http.Client.Timeout` 早于 context deadline 到期，会覆盖调用方意图中的超时时间，导致行为与预期不一致。

## 完整示例

### 创建客户端

```go
import "github.com/lerity-yao/czt-contrib/aliyun/oss"

client := oss.MustNewClient(oss.Conf{
    Endpoint:        "oss-cn-hangzhou.aliyuncs.com",
    Bucket:          "my-bucket",
    AccessKeyId:     "your-access-key-id",
    AccessKeySecret: "your-access-key-secret",
})
```

> 本地 MinIO 测试：Endpoint 加 `http://` 前缀即可走 HTTP。
>
> ```go
> oss.Conf{Endpoint: "http://127.0.0.1:9000", Bucket: "test", ...}
> ```

### PutObject：上传对象

```go
// 1. 简单上传
err := client.PutObject(ctx, "photos/test.jpg", fileBytes)

// 2. 指定 Content-Type + 自定义元数据
err := client.PutObject(ctx, "photos/test.jpg", fileBytes,
    oss.WithContentType("image/jpeg"),
    oss.WithMeta("author", "tom"),
)
```

### GetObject：下载对象

```go
// 流式下载，调用方负责关闭 Reader
rc, err := client.GetObject(ctx, "photos/test.jpg")
if err != nil {
    return err
}
defer rc.Close()

io.Copy(os.Stdout, rc)
```

### DeleteObject：删除对象

```go
err := client.DeleteObject(ctx, "photos/test.jpg")
```

### HeadObject：获取元信息

```go
meta, err := client.HeadObject(ctx, "photos/test.jpg")
// meta.Size         // 文件大小（字节）
// meta.ContentType  // "image/jpeg"
// meta.ETag         // "d41d8cd98f00b204e9800998ecf8427e"
// meta.LastModified // "Wed, 14 Jun 2026 12:00:00 GMT"
// meta.Metadata     // map["x-oss-meta-author"]"tom"
```

### CopyObject：拷贝对象

```go
// 同 Bucket 内拷贝
err := client.CopyObject(ctx, "photos/copy.jpg", "photos/test.jpg")
```

### ListObjects：列举对象

```go
// 1. 按前缀列举
result, err := client.ListObjects(ctx,
    oss.WithPrefix("photos/"),
    oss.WithMaxKeys(50),
)
// result.IsTruncated    // 是否还有更多
// result.NextMarker     // 下一页起始 Key
// result.Contents       // []ObjectInfo
//   .Key / .Size / .ETag / .LastModified / .Type / .StorageClass

// 2. 分页列举（模拟目录结构）
result, err := client.ListObjects(ctx,
    oss.WithPrefix("photos/"),
    oss.WithDelimiter("/"),
)
// result.Contents       // 当前层级的文件
// result.CommonPrefixes // 子目录（[]CommonPrefix{.Prefix}）
```

### Do：原始签名请求

```go
// 获取对象 ACL（子资源）
resp, err := client.Do(ctx, http.MethodGet, "photos/test.jpg",
    map[string]string{}, nil)
// OSS 子资源通过 query 参数传递，Do 方法会自动处理签名
// 但当前版本 Do 不直接支持 query 参数，可改用 ListObjects 的方式
defer resp.Body.Close()

// 设置自定义 header
resp, err = client.Do(ctx, http.MethodPut, "photos/test.jpg",
    map[string]string{
        "Content-Type":       "image/jpeg",
        "x-oss-object-acl":   "public-read",
        "x-oss-meta-author":  "tom",
    },
    fileBytes,
)
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

client := oss.MustNewClient(conf, oss.WithClient(cli))
```

### 在 go-zero 中使用

```go
// internal/config/config.go
type Config struct {
    rest.RestConf
    OSS oss.Conf
}
```

```yaml
# etc/config.yaml
Name: api
Host: 0.0.0.0
Port: 8888

OSS:
  Endpoint: oss-cn-hangzhou.aliyuncs.com
  Bucket: my-bucket
  AccessKeyId: your-access-key-id
  AccessKeySecret: your-access-key-secret
```

```go
// internal/svc/servicecontext.go
type ServiceContext struct {
    Config config.Config
    OSS    oss.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
    return &ServiceContext{
        Config: c,
        OSS:    oss.MustNewClient(c.OSS),
    }
}
```

```go
// internal/logic/uploadlogic.go
func (l *UploadLogic) Upload(req *types.UploadReq) error {
    err := l.svcCtx.OSS.PutObject(l.ctx, req.Key, req.Content,
        oss.WithContentType(req.ContentType),
    )
    return err
}
```

## 更新日志

查看 [CHANGELOG.md](./CHANGELOG.md)
