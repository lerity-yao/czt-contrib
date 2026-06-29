# minio

[English](./readme-en.md) | 中文

基于 [minio-go v7](https://github.com/minio/minio-go) 封装的 MinIO Go 客户端 SDK，集成 [go-zero](https://github.com/zeromicro/go-zero) 生态（熔断、追踪、指标、慢日志），内置 P2C 负载均衡和写后读亲和。

## Features

- 🔀 **P2C 负载均衡** — 多节点直连，自动选择最优节点
- 🔁 **写后读亲和** — 5s TTL，解决多节点复制延迟
- 🛡️ **熔断保护** — 5xx 和网络错误自动熔断
- 📊 **全链路可观测** — Prometheus 指标 + OpenTelemetry 追踪 + 慢日志
- 🔄 **自动故障转移** — 首选节点失败后 shuffle 确定性遍历剩余节点
- 🔐 **V2/V4 双签名** — 同时支持 AWS Signature V2 和 V4
- 📦 **双层 API** — 便捷方法 + 原子操作 + RawClient

## 安装

```bash
go get github.com/lerity-yao/czt-contrib/minio@v0.0.1
```

## 配置

### Conf

| 字段 | 类型 | 必传/可选 | 默认值 | 说明 |
|------|------|-----------|--------|------|
| `Endpoints` | []string | 必传 | — | MinIO 服务端地址列表，如 `["192.168.1.10:9000", "192.168.1.11:9000"]` |
| `AccessKeyID` | string | 必传 | — | 访问密钥 ID |
| `SecretAccessKey` | string | 必传 | — | 访问密钥 Secret |
| `UseSSL` | bool | 可选 | `false` | 是否启用 HTTPS；直连 IP 场景通常为 false |
| `Region` | string | 可选 | `""` | 服务端区域，如 `us-east-1` |
| `SignatureVersion` | string | 可选 | `v4` | 签名算法版本；支持 `v2` / `v4` |
| `SlowThreshold` | int64 | 可选 | `1000` | 慢请求日志阈值（毫秒）；设为 0 禁用慢日志 |

`NewClient` 创建时自动调用 `Validate()` 校验以上字段。

> `Endpoints`、`AccessKeyID`、`SecretAccessKey` 为必传字段，不带 json tag（go-zero 规范）。`UseSSL`、`Region`、`SignatureVersion`、`SlowThreshold` 支持 go-zero `conf.MustLoad` 的 `optional` / `default` tag，可在 YAML 中省略。

## API 参考

### 构造函数

| 函数 | 签名 | 说明 |
|------|------|------|
| `NewClient` | `func NewClient(c Conf, opts ...ClientOption) (Client, error)` | 创建 Client，校验失败返回 error |
| `MustNewClient` | `func MustNewClient(c Conf, opts ...ClientOption) Client` | 创建 Client，校验失败 panic |

### ClientOption

| Option | 参数 | 说明 |
|--------|------|------|
| `WithTransport` | `http.RoundTripper` | 注入自定义 base Transport（超时、TLS、连接池等）；可观测 Transport 会包装此 Transport |

### Option（上传用）

| Option | 参数 | 说明 |
|--------|------|------|
| `WithContentType` | `string` | 设置上传对象的 Content-Type；默认 `application/octet-stream` |
| `WithMetadata` | `map[string]string` | 设置自定义元数据 |
| `WithStorageClass` | `string` | 设置存储类别 |
| `WithPartSize` | `uint64` | 设置分片上传大小；默认 16 MiB |

### PresignedOption（预签名 URL 用）

| Option | 参数 | 说明 |
|--------|------|------|
| `WithResponseContentDisposition` | `string` | 设置预签名 URL 的 `response-content-disposition`；`"inline"` 浏览器预览，`"attachment"` 强制下载 |
| `WithResponseContentType` | `string` | 覆盖预签名 URL 的 `response-content-type` |

### Client 接口方法

#### 便捷方法

| 方法 | 签名 | 说明 |
|------|------|------|
| `UploadFile` | `UploadFile(ctx, bucket, key, filePath string, opts ...Option) (*UploadInfo, error)` | 上传本地文件，自动分片 |
| `UploadReader` | `UploadReader(ctx, bucket, key string, reader io.Reader, size int64, opts ...Option) (*UploadInfo, error)` | 从 Reader 上传 |
| `Download` | `Download(ctx, bucket, key string, opts ...Option) (io.ReadCloser, error)` | 下载对象，返回流；调用方需 Close |
| `Delete` | `Delete(ctx, bucket, key string) error` | 删除对象 |
| `Exists` | `Exists(ctx, bucket, key string) (bool, error)` | 判断对象是否存在 |
| `GetPresignedDownloadURL` | `GetPresignedDownloadURL(ctx, bucket, key string, expiry time.Duration, opts ...PresignedOption) (string, error)` | 生成预签名下载 URL |
| `GetPresignedUploadURL` | `GetPresignedUploadURL(ctx, bucket, key string, expiry time.Duration) (string, error)` | 生成预签名上传 URL |

#### 原子操作

| 方法 | 签名 | 说明 |
|------|------|------|
| `PutObject` | `PutObject(ctx, bucket, key string, reader io.Reader, size int64, opts miniogo.PutObjectOptions) (*UploadInfo, error)` | 完全控制上传参数 |
| `GetObject` | `GetObject(ctx, bucket, key string, opts miniogo.GetObjectOptions) (*miniogo.Object, error)` | 完全控制下载参数 |
| `StatObject` | `StatObject(ctx, bucket, key string, opts miniogo.StatObjectOptions) (*ObjectInfo, error)` | 获取对象元数据 |
| `RemoveObject` | `RemoveObject(ctx, bucket, key string, opts miniogo.RemoveObjectOptions) error` | 完全控制删除参数 |
| `CopyObject` | `CopyObject(ctx context.Context, dst miniogo.CopyDestOptions, src miniogo.CopySrcOptions) (*UploadInfo, error)` | 复制对象 |
| `ListObjects` | `ListObjects(ctx, bucket string, opts miniogo.ListObjectsOptions) <-chan miniogo.ObjectInfo` | 列举对象，返回 channel |

#### Bucket 管理

| 方法 | 签名 | 说明 |
|------|------|------|
| `MakeBucket` | `MakeBucket(ctx, bucket string, opts miniogo.MakeBucketOptions) error` | 创建 Bucket |
| `RemoveBucket` | `RemoveBucket(ctx, bucket string) error` | 删除空 Bucket |
| `ListBuckets` | `ListBuckets(ctx) ([]BucketInfo, error)` | 列举所有 Bucket |
| `SetBucketPolicy` | `SetBucketPolicy(ctx, bucket, policy string) error` | 设置 Bucket 策略 |
| `GetBucketPolicy` | `GetBucketPolicy(ctx, bucket string) (string, error)` | 获取 Bucket 策略 |

#### 原始客户端

| 方法 | 签名 | 说明 |
|------|------|------|
| `RawClient` | `RawClient() *miniogo.Client` | 获取 P2C 选中的底层 minio-go 客户端 |
| `RawClients` | `RawClients() []*miniogo.Client` | 获取全部底层 minio-go 客户端 |

### 返回类型

#### UploadInfo

| 字段 | 类型 | 说明 |
|------|------|------|
| `Key` | string | 对象键 |
| `ETag` | string | 对象 ETag |
| `Size` | int64 | 对象大小（字节） |
| `VersionID` | string | 版本 ID（开启版本控制时） |

#### ObjectInfo

| 字段 | 类型 | 说明 |
|------|------|------|
| `Key` | string | 对象键 |
| `Size` | int64 | 对象大小（字节） |
| `ContentType` | string | 内容类型 |
| `ETag` | string | 对象 ETag |
| `LastModified` | time.Time | 最后修改时间 |
| `Metadata` | map[string]string | 用户自定义元数据 |

#### BucketInfo

| 字段 | 类型 | 说明 |
|------|------|------|
| `Name` | string | Bucket 名称 |
| `CreationDate` | time.Time | 创建时间 |

### 错误处理

SDK 将 minio-go 的 `ErrorResponse` 包装为自定义 `Error` 类型，提供结构化错误信息：

| 字段 | 类型 | 说明 |
|------|------|------|
| `Code` | string | MinIO 错误码，如 `NoSuchKey`、`NoSuchBucket` |
| `Message` | string | 错误描述 |
| `BucketName` | string | 相关 Bucket |
| `Key` | string | 相关对象键 |
| `StatusCode` | int | HTTP 状态码 |
| `Cause` | error | 原始错误（支持 `errors.Unwrap`） |

非 `ErrorResponse` 类型的错误（如网络错误）会原样返回。

```go
import (
    "errors"

    "github.com/lerity-yao/czt-contrib/minio"
)

_, err := client.Download(ctx, "my-bucket", "not-exist.txt")
if err != nil {
    var minioErr *minio.Error
    if errors.As(err, &minioErr) {
        fmt.Printf("code=%s status=%d bucket=%s key=%s\n",
            minioErr.Code, minioErr.StatusCode, minioErr.BucketName, minioErr.Key)
    }
}
```

## 高级指南

### 负载均衡机制

SDK 内置 **P2C (Power of Two Choices)** 负载均衡算法，灵感来自 go-zero 的 p2c 实现：

1. **随机取两个节点**，比较两者的负载分数
2. **负载分数** = `sqrt(EWMA延迟 + 1) × (并发数 + 1)`
3. 选择负载分数更低的节点
4. **EWMA 衰减窗口** 10 秒，自动平滑延迟抖动
5. 若某节点超过 1 秒未被选中（`forcePick`），强制选择以避免饥饿

> 单节点直接选择；双节点直接比较；3 个及以上节点随机取两个比较。

### 写后读亲和

解决多节点部署下数据复制延迟导致"写完立即读不到"的问题：

1. **写操作成功后**，将 `bucket/key → nodeIndex` 写入本地 TimingWheel 缓存
2. **后续读操作**先查亲和缓存，命中则路由到同一节点
3. **TTL 为 5 秒**，超过后缓存自动过期，恢复 P2C 选择
4. 亲和读失败时，惩罚该节点负载并回退到 P2C 故障转移

适用方法：`UploadFile`、`UploadReader`、`PutObject`、`CopyObject` 写入后设置亲和；`Download`、`Exists`、`StatObject`、`GetObject` 读取时查询亲和。

### 故障转移

所有非流式操作均支持自动故障转移：

1. **首选节点** 通过 P2C 算法选出
2. 若发生**网络级错误**（连接拒绝、连接重置、DNS 解析失败、超时等），自动故障转移
3. **Shuffle 剩余节点**，确定性遍历，直到成功或全部失败
4. **非网络错误**（如 4xx 业务错误）不触发故障转移，直接返回

> 流式操作（`Download`、`GetObject`、`ListObjects`）因返回流/channel，不支持自动故障转移。

### 可观测性

#### Prometheus 指标

| 指标名 | 类型 | 维度 | 说明 |
|--------|------|------|------|
| `minio_client_requests_duration_ms` | Histogram | `method`, `bucket`, `endpoint` | 请求耗时（毫秒），分桶 5/10/25/50/100/250/500/1000/2500 |
| `minio_client_requests_code_total` | Counter | `method`, `bucket`, `code`, `endpoint` | 按状态码计数；`0` 表示无响应（网络错误） |
| `minio_client_requests_failover_total` | Counter | `endpoint` | 首选节点失败触发故障转移的次数 |
| `minio_client_affinity_hit_total` | Counter | `bucket` | 写后读亲和缓存命中次数 |
| `minio_client_affinity_miss_total` | Counter | `bucket` | 写后读亲和缓存未命中次数 |
| `minio_client_breaker_trip_total` | Counter | `endpoint` | 熔断器跳闸次数 |

#### OpenTelemetry 追踪

每个 HTTP 请求自动创建 Span：

- **Tracer 名称**：`github.com/lerity-yao/czt-contrib/minio`
- **Span 名称**：`minio-client.<HTTP方法>`（如 `minio-client.PUT`）
- **Span Kind**：`SpanKindClient`
- 自动注入 trace context 到请求 Header（W3C Trace Context 传播）

#### 慢日志

当请求耗时超过 `SlowThreshold`（默认 1000ms）或发生错误时，自动记录慢日志：

```
[minio] PUT /my-bucket/key.txt bucket=my-bucket duration=2500ms err=<nil>
```

设置 `SlowThreshold: 0` 禁用慢日志。

#### 熔断器

- 每个 endpoint 独立熔断器，命名为 `minio:<endpoint>`
- 5xx 响应和网络错误视为失败
- 熔断触发时返回 `breaker.ErrServiceUnavailable`，并递增 `minio_client_breaker_trip_total` 指标

### 在 go-zero 中使用

```go
// internal/config/config.go
package config

import (
    "github.com/lerity-yao/czt-contrib/minio"
    "github.com/zeromicro/go-zero/rest"
)

type Config struct {
    rest.RestConf
    Minio minio.Conf
}
```

```yaml
# etc/config.yaml
Name: file-api
Host: 0.0.0.0
Port: 8888

Minio:
  Endpoints:
    - 192.168.1.10:9000
    - 192.168.1.11:9000
    - 192.168.1.12:9000
    - 192.168.1.13:9000
  AccessKeyID: your-access-key
  SecretAccessKey: your-secret-key
  # 以下可省略，使用默认值
  # UseSSL: false
  # SignatureVersion: v4
  # SlowThreshold: 1000
```

```go
// internal/svc/servicecontext.go
package svc

import (
    "github.com/lerity-yao/czt-contrib/minio"
    "your-project/internal/config"
)

type ServiceContext struct {
    Config config.Config
    Minio  minio.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
    return &ServiceContext{
        Config: c,
        Minio:  minio.MustNewClient(c.Minio),
    }
}
```

```go
// internal/logic/uploadlogic.go
package logic

import (
    "context"
    "time"

    "github.com/lerity-yao/czt-contrib/minio"
    "your-project/internal/svc"
)

type UploadLogic struct {
    ctx    context.Context
    svcCtx *svc.ServiceContext
}

func (l *UploadLogic) Upload(filePath string) (*minio.UploadInfo, error) {
    // 上传文件
    info, err := l.svcCtx.Minio.UploadFile(l.ctx, "my-bucket", "uploads/test.png", filePath,
        minio.WithContentType("image/png"),
    )
    if err != nil {
        return nil, err
    }

    // 生成预签名下载 URL（浏览器预览）
    url, err := l.svcCtx.Minio.GetPresignedDownloadURL(l.ctx, "my-bucket", "uploads/test.png",
        10*time.Minute,
        minio.WithResponseContentDisposition("inline"),
    )
    if err != nil {
        return nil, err
    }
    _ = url // 返回给前端

    return info, nil
}
```

### 独立脚本使用

```go
package main

import (
    "context"
    "fmt"
    "os"
    "time"

    "github.com/lerity-yao/czt-contrib/minio"
)

func main() {
    client := minio.MustNewClient(minio.Conf{
        Endpoints:      []string{"192.168.1.10:9000"},
        AccessKeyID:    "your-access-key",
        SecretAccessKey: "your-secret-key",
    })

    ctx := context.Background()

    // 上传文件
    info, err := client.UploadFile(ctx, "my-bucket", "test.txt", "/tmp/test.txt")
    if err != nil {
        fmt.Fprintf(os.Stderr, "upload failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Printf("uploaded: key=%s etag=%s size=%d\n", info.Key, info.ETag, info.Size)

    // 检查对象是否存在
    exists, err := client.Exists(ctx, "my-bucket", "test.txt")
    if err != nil {
        fmt.Fprintf(os.Stderr, "exists check failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("exists:", exists)

    // 下载
    reader, err := client.Download(ctx, "my-bucket", "test.txt")
    if err != nil {
        fmt.Fprintf(os.Stderr, "download failed: %v\n", err)
        os.Exit(1)
    }
    defer reader.Close()

    // 生成预签名 URL
    url, err := client.GetPresignedDownloadURL(ctx, "my-bucket", "test.txt", 1*time.Hour)
    if err != nil {
        fmt.Fprintf(os.Stderr, "presign failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("presigned URL:", url)

    // 删除
    if err := client.Delete(ctx, "my-bucket", "test.txt"); err != nil {
        fmt.Fprintf(os.Stderr, "delete failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("deleted")
}
```

## API 使用案例

以下示例假设 `client` 已通过 `minio.MustNewClient(conf)` 创建，`ctx` 为有效的 `context.Context`。

### 便捷方法

#### UploadFile

```go
// 上传本地文件，设置 Content-Type 和自定义元数据
info, err := client.UploadFile(ctx, "my-bucket", "docs/report.pdf", "/tmp/report.pdf",
    minio.WithContentType("application/pdf"),
    minio.WithMetadata(map[string]string{"author": "test"}),
)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("uploaded: key=%s size=%d\n", info.Key, info.Size)
```

#### UploadReader

```go
import "strings"

// 从 Reader 上传（如内存数据、HTTP 请求体）
data := strings.NewReader("hello world")
info, err := client.UploadReader(ctx, "my-bucket", "notes/hello.txt", data, int64(data.Len()),
    minio.WithContentType("text/plain"),
    minio.WithStorageClass("STANDARD"),
)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("uploaded: key=%s etag=%s\n", info.Key, info.ETag)
```

#### Download

```go
// 下载对象，调用方需关闭 reader
reader, err := client.Download(ctx, "my-bucket", "notes/hello.txt")
if err != nil {
    log.Fatal(err)
}
defer reader.Close()

io.Copy(os.Stdout, reader)
```

#### Delete

```go
// 删除对象
if err := client.Delete(ctx, "my-bucket", "notes/hello.txt"); err != nil {
    log.Fatal(err)
}
```

#### Exists

```go
// 判断对象是否存在
exists, err := client.Exists(ctx, "my-bucket", "notes/hello.txt")
if err != nil {
    log.Fatal(err)
}
fmt.Println("exists:", exists)
```

#### GetPresignedDownloadURL

```go
import "time"

// 生成预签名下载 URL（浏览器内预览图片）
url, err := client.GetPresignedDownloadURL(ctx, "my-bucket", "images/photo.jpg",
    10*time.Minute,
    minio.WithResponseContentDisposition("inline"),
    minio.WithResponseContentType("image/jpeg"),
)
if err != nil {
    log.Fatal(err)
}
fmt.Println("preview URL:", url)
```

#### GetPresignedUploadURL

```go
import "time"

// 生成预签名上传 URL，前端可直接 PUT 上传
uploadURL, err := client.GetPresignedUploadURL(ctx, "my-bucket", "uploads/avatar.png", 15*time.Minute)
if err != nil {
    log.Fatal(err)
}
fmt.Println("upload URL:", uploadURL)
// 前端使用: fetch(uploadURL, { method: 'PUT', body: file })
```

### 原子操作

#### PutObject

```go
import (
    "strings"

    miniogo "github.com/minio/minio-go/v7"
)

// 完全控制上传参数
data := strings.NewReader("{\"key\": \"value\"}")
info, err := client.PutObject(ctx, "my-bucket", "data/config.json", data, int64(data.Len()),
    miniogo.PutObjectOptions{
        ContentType:  "application/json",
        UserMetadata: map[string]string{"version": "1"},
    },
)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("put: key=%s etag=%s\n", info.Key, info.ETag)
```

#### GetObject

```go
import miniogo "github.com/minio/minio-go/v7"

// 完全控制下载参数（如指定 Range 下载）
obj, err := client.GetObject(ctx, "my-bucket", "data/config.json", miniogo.GetObjectOptions{})
if err != nil {
    log.Fatal(err)
}
defer obj.Close()

io.Copy(os.Stdout, obj)
```

#### StatObject

```go
import miniogo "github.com/minio/minio-go/v7"

// 获取对象元数据（大小、Content-Type、修改时间等）
objInfo, err := client.StatObject(ctx, "my-bucket", "data/config.json", miniogo.StatObjectOptions{})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("key=%s size=%d contentType=%s lastModified=%s\n",
    objInfo.Key, objInfo.Size, objInfo.ContentType, objInfo.LastModified)
```

#### RemoveObject

```go
import miniogo "github.com/minio/minio-go/v7"

// 完全控制删除参数（如指定版本 ID）
err := client.RemoveObject(ctx, "my-bucket", "data/config.json", miniogo.RemoveObjectOptions{
    VersionID: "specific-version-id",
})
if err != nil {
    log.Fatal(err)
}
```

#### CopyObject

```go
import miniogo "github.com/minio/minio-go/v7"

// 复制对象到另一位置
info, err := client.CopyObject(ctx,
    miniogo.CopyDestOptions{
        Bucket: "my-bucket",
        Object: "backup/config.json",
    },
    miniogo.CopySrcOptions{
        Bucket: "my-bucket",
        Object: "data/config.json",
    },
)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("copied: key=%s etag=%s\n", info.Key, info.ETag)
```

#### ListObjects

```go
import miniogo "github.com/minio/minio-go/v7"

// 列举指定前缀的所有对象
for obj := range client.ListObjects(ctx, "my-bucket", miniogo.ListObjectsOptions{
    Prefix:    "data/",
    Recursive: true,
}) {
    if obj.Err != nil {
        log.Fatal(obj.Err)
    }
    fmt.Printf("key=%s size=%d\n", obj.Key, obj.Size)
}
```

### Bucket 管理

#### MakeBucket

```go
import miniogo "github.com/minio/minio-go/v7"

// 创建 Bucket
err := client.MakeBucket(ctx, "new-bucket", miniogo.MakeBucketOptions{
    Region: "us-east-1",
})
if err != nil {
    log.Fatal(err)
}
```

#### RemoveBucket

```go
// 删除空 Bucket（需确保 Bucket 内无对象）
err := client.RemoveBucket(ctx, "empty-bucket")
if err != nil {
    log.Fatal(err)
}
```

#### ListBuckets

```go
// 列举所有 Bucket
buckets, err := client.ListBuckets(ctx)
if err != nil {
    log.Fatal(err)
}
for _, b := range buckets {
    fmt.Printf("name=%s created=%s\n", b.Name, b.CreationDate)
}
```

#### SetBucketPolicy

```go
// 设置 Bucket 为公开只读策略
policy := `{
    "Version": "2012-10-17",
    "Statement": [{
        "Effect": "Allow",
        "Principal": {"AWS": ["*"]},
        "Action": ["s3:GetObject"],
        "Resource": ["arn:aws:s3:::my-bucket/*"]
    }]
}`
err := client.SetBucketPolicy(ctx, "my-bucket", policy)
if err != nil {
    log.Fatal(err)
}
```

#### GetBucketPolicy

```go
// 获取 Bucket 当前策略
policy, err := client.GetBucketPolicy(ctx, "my-bucket")
if err != nil {
    log.Fatal(err)
}
fmt.Println("policy:", policy)
```

### 原始客户端

#### RawClient

```go
import miniogo "github.com/minio/minio-go/v7"

// 获取 P2C 选中的底层 minio-go 客户端，执行 SDK 未封装的操作
raw := client.RawClient()

// 例：使用原始客户端设置 Bucket 生命周期
import "github.com/minio/minio-go/v7/pkg/lifecycle"

config := lifecycle.NewConfiguration()
config.Rules = []lifecycle.Rule{
    {
        ID:     "expire-old",
        Status: "Enabled",
        Expiration: lifecycle.Expiration{
            Days: 30,
        },
    },
}
err := raw.SetBucketLifecycle(ctx, "my-bucket", config)
if err != nil {
    log.Fatal(err)
}
```

#### RawClients

```go
// 获取全部底层客户端，用于自定义巡检或批量操作
clients := client.RawClients()
for i, raw := range clients {
    alive, _ := raw.HealthCheck(10 * time.Second)
    fmt.Printf("node[%d] healthy=%v\n", i, alive)
}
```

### WithTransport 示例

```go
import (
    "crypto/tls"
    "net"
    "net/http"
    "time"

    "github.com/lerity-yao/czt-contrib/minio"
)

// 注入自定义 Transport（如自定义超时、TLS 配置）
customTransport := &http.Transport{
    TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
    MaxIdleConnsPerHost: 50,
    IdleConnTimeout:     90 * time.Second,
    DialContext: (&net.Dialer{
        Timeout: 5 * time.Second,
    }).DialContext,
}

client := minio.MustNewClient(conf, minio.WithTransport(customTransport))
```

### WithPartSize 示例

```go
// 大文件上传时设置分片大小为 64 MiB
info, err := client.UploadFile(ctx, "my-bucket", "bigfile.bin", "/data/bigfile.bin",
    minio.WithPartSize(64*1024*1024),
)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("uploaded: key=%s size=%d\n", info.Key, info.Size)
```

## Changelog

See [CHANGELOG.md](./CHANGELOG.md)
