# minio

English | [中文](./readme-cn.md)

A MinIO Go client SDK built on [minio-go v7](https://github.com/minio/minio-go), integrated with the [go-zero](https://github.com/zeromicro/go-zero) ecosystem (circuit breaker, tracing, metrics, slow log), featuring built-in P2C load balancing and write-after-read affinity.

## Features

- 🔀 **P2C Load Balancing** — Direct multi-node connection, automatically selects the optimal node
- 🔁 **Write-After-Read Affinity** — 5s TTL, solves multi-node replication delay
- 🛡️ **Circuit Breaker Protection** — Auto-trips on 5xx and network errors
- 📊 **Full Observability** — Prometheus metrics + OpenTelemetry tracing + slow log
- 🔄 **Automatic Failover** — Shuffle deterministic traversal of remaining nodes after preferred node failure
- 🔐 **V2/V4 Dual Signature** — Supports both AWS Signature V2 and V4
- 📦 **Dual-Layer API** — Convenience methods + low-level operations + RawClient

## Installation

```bash
go get github.com/lerity-yao/czt-contrib/minio@v0.0.1
```

## Configuration

### Conf

| Field | Type | Required/Optional | Default | Description |
|-------|------|-------------------|---------|-------------|
| `Endpoints` | []string | Required | — | MinIO server address list, e.g. `["192.168.1.10:9000", "192.168.1.11:9000"]` |
| `AccessKeyID` | string | Required | — | Access key ID |
| `SecretAccessKey` | string | Required | — | Access key secret |
| `UseSSL` | bool | Optional | `false` | Whether to enable HTTPS; typically false for direct IP connections |
| `Region` | string | Optional | `""` | Server region, e.g. `us-east-1` |
| `SignatureVersion` | string | Optional | `v4` | Signature algorithm version; supports `v2` / `v4` |
| `SlowThreshold` | int64 | Optional | `1000` | Slow request log threshold (milliseconds); set to 0 to disable slow log |

`NewClient` automatically calls `Validate()` to validate the above fields during creation.

> `Endpoints`, `AccessKeyID`, `SecretAccessKey` are required fields without json tags (go-zero convention). `UseSSL`, `Region`, `SignatureVersion`, `SlowThreshold` support go-zero `conf.MustLoad`'s `optional` / `default` tags and can be omitted from YAML.

## API Reference

### Constructors

| Function | Signature | Description |
|----------|-----------|-------------|
| `NewClient` | `func NewClient(c Conf, opts ...ClientOption) (Client, error)` | Creates a Client; returns error on validation failure |
| `MustNewClient` | `func MustNewClient(c Conf, opts ...ClientOption) Client` | Creates a Client; panics on validation failure |

### ClientOption

| Option | Parameter | Description |
|--------|-----------|-------------|
| `WithTransport` | `http.RoundTripper` | Injects a custom base Transport (timeout, TLS, connection pool, etc.); the instrumented Transport wraps this Transport |

### Option (for uploads)

| Option | Parameter | Description |
|--------|-----------|-------------|
| `WithContentType` | `string` | Sets the Content-Type of the uploaded object; defaults to `application/octet-stream` |
| `WithMetadata` | `map[string]string` | Sets custom metadata |
| `WithStorageClass` | `string` | Sets the storage class |
| `WithPartSize` | `uint64` | Sets the multipart upload part size; defaults to 16 MiB |

### PresignedOption (for presigned URLs)

| Option | Parameter | Description |
|--------|-----------|-------------|
| `WithResponseContentDisposition` | `string` | Sets `response-content-disposition` for the presigned URL; `"inline"` for browser preview, `"attachment"` for forced download |
| `WithResponseContentType` | `string` | Overrides `response-content-type` for the presigned URL |

### Client Interface Methods

#### Convenience Methods

| Method | Signature | Description |
|--------|-----------|-------------|
| `UploadFile` | `UploadFile(ctx, bucket, key, filePath string, opts ...Option) (*UploadInfo, error)` | Uploads a local file with automatic multipart |
| `UploadReader` | `UploadReader(ctx, bucket, key string, reader io.Reader, size int64, opts ...Option) (*UploadInfo, error)` | Uploads from a Reader |
| `Download` | `Download(ctx, bucket, key string, opts ...Option) (io.ReadCloser, error)` | Downloads an object, returns a stream; caller must Close |
| `Delete` | `Delete(ctx, bucket, key string) error` | Deletes an object |
| `Exists` | `Exists(ctx, bucket, key string) (bool, error)` | Checks whether an object exists |
| `GetPresignedDownloadURL` | `GetPresignedDownloadURL(ctx, bucket, key string, expiry time.Duration, opts ...PresignedOption) (string, error)` | Generates a presigned download URL |
| `GetPresignedUploadURL` | `GetPresignedUploadURL(ctx, bucket, key string, expiry time.Duration) (string, error)` | Generates a presigned upload URL |

#### Low-Level Operations

| Method | Signature | Description |
|--------|-----------|-------------|
| `PutObject` | `PutObject(ctx, bucket, key string, reader io.Reader, size int64, opts miniogo.PutObjectOptions) (*UploadInfo, error)` | Full control over upload parameters |
| `GetObject` | `GetObject(ctx, bucket, key string, opts miniogo.GetObjectOptions) (*miniogo.Object, error)` | Full control over download parameters |
| `StatObject` | `StatObject(ctx, bucket, key string, opts miniogo.StatObjectOptions) (*ObjectInfo, error)` | Gets object metadata |
| `RemoveObject` | `RemoveObject(ctx, bucket, key string, opts miniogo.RemoveObjectOptions) error` | Full control over delete parameters |
| `CopyObject` | `CopyObject(ctx context.Context, dst miniogo.CopyDestOptions, src miniogo.CopySrcOptions) (*UploadInfo, error)` | Copies an object |
| `ListObjects` | `ListObjects(ctx, bucket string, opts miniogo.ListObjectsOptions) <-chan miniogo.ObjectInfo` | Lists objects, returns a channel |

#### Bucket Management

| Method | Signature | Description |
|--------|-----------|-------------|
| `MakeBucket` | `MakeBucket(ctx, bucket string, opts miniogo.MakeBucketOptions) error` | Creates a bucket |
| `RemoveBucket` | `RemoveBucket(ctx, bucket string) error` | Removes an empty bucket |
| `ListBuckets` | `ListBuckets(ctx) ([]BucketInfo, error)` | Lists all buckets |
| `SetBucketPolicy` | `SetBucketPolicy(ctx, bucket, policy string) error` | Sets bucket policy |
| `GetBucketPolicy` | `GetBucketPolicy(ctx, bucket string) (string, error)` | Gets bucket policy |

#### Raw Client

| Method | Signature | Description |
|--------|-----------|-------------|
| `RawClient` | `RawClient() *miniogo.Client` | Gets the P2C-selected underlying minio-go client |
| `RawClients` | `RawClients() []*miniogo.Client` | Gets all underlying minio-go clients |

### Return Types

#### UploadInfo

| Field | Type | Description |
|-------|------|-------------|
| `Key` | string | Object key |
| `ETag` | string | Object ETag |
| `Size` | int64 | Object size (bytes) |
| `VersionID` | string | Version ID (when versioning is enabled) |

#### ObjectInfo

| Field | Type | Description |
|-------|------|-------------|
| `Key` | string | Object key |
| `Size` | int64 | Object size (bytes) |
| `ContentType` | string | Content type |
| `ETag` | string | Object ETag |
| `LastModified` | time.Time | Last modified time |
| `Metadata` | map[string]string | User-defined metadata |

#### BucketInfo

| Field | Type | Description |
|-------|------|-------------|
| `Name` | string | Bucket name |
| `CreationDate` | time.Time | Creation time |

### Error Handling

The SDK wraps minio-go's `ErrorResponse` into a custom `Error` type, providing structured error information:

| Field | Type | Description |
|-------|------|-------------|
| `Code` | string | MinIO error code, e.g. `NoSuchKey`, `NoSuchBucket` |
| `Message` | string | Error description |
| `BucketName` | string | Related bucket |
| `Key` | string | Related object key |
| `StatusCode` | int | HTTP status code |
| `Cause` | error | Original error (supports `errors.Unwrap`) |

Non-`ErrorResponse` errors (e.g. network errors) are returned as-is.

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

## Advanced Guide

### Load Balancing

The SDK has a built-in **P2C (Power of Two Choices)** load balancing algorithm, inspired by go-zero's p2c implementation:

1. **Randomly pick two nodes** and compare their load scores
2. **Load score** = `sqrt(EWMA latency + 1) × (inflight + 1)`
3. Select the node with the lower load score
4. **EWMA decay window** of 10 seconds, automatically smoothing latency jitter
5. If a node has not been selected for more than 1 second (`forcePick`), it is force-selected to avoid starvation

> Single node: direct selection; two nodes: direct comparison; 3+ nodes: randomly pick two and compare.

### Write-After-Read Affinity

Solves the "write then immediately read fails" problem caused by replication delay in multi-node deployments:

1. **After a successful write**, records `bucket/key → nodeIndex` in a local TimingWheel cache
2. **Subsequent reads** first query the affinity cache; on hit, route to the same node
3. **TTL is 5 seconds**; after expiration, the cache auto-expires and reverts to P2C selection
4. If the affinity read fails, the node's load is penalized and falls back to P2C failover

Applicable methods: `UploadFile`, `UploadReader`, `PutObject`, `CopyObject` set affinity after write; `Download`, `Exists`, `StatObject`, `GetObject` query affinity on read.

### Failover

All non-streaming operations support automatic failover:

1. **Preferred node** is selected via the P2C algorithm
2. If a **network-level error** occurs (connection refused, connection reset, DNS resolution failure, timeout, etc.), automatic failover is triggered
3. **Shuffle remaining nodes** with deterministic traversal until success or all nodes fail
4. **Non-network errors** (e.g. 4xx business errors) do not trigger failover and are returned directly

> Streaming operations (`Download`, `GetObject`, `ListObjects`) return streams/channels and do not support automatic failover.

### Observability

#### Prometheus Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `minio_client_requests_duration_ms` | Histogram | `method`, `bucket`, `endpoint` | Request duration (ms), buckets: 5/10/25/50/100/250/500/1000/2500 |
| `minio_client_requests_code_total` | Counter | `method`, `bucket`, `code`, `endpoint` | Count by status code; `0` means no response (network error) |
| `minio_client_requests_failover_total` | Counter | `endpoint` | Number of failovers triggered by preferred node failure |
| `minio_client_affinity_hit_total` | Counter | `bucket` | Write-after-read affinity cache hits |
| `minio_client_affinity_miss_total` | Counter | `bucket` | Write-after-read affinity cache misses |
| `minio_client_breaker_trip_total` | Counter | `endpoint` | Circuit breaker trip count |

#### OpenTelemetry Tracing

Each HTTP request automatically creates a Span:

- **Tracer name**: `github.com/lerity-yao/czt-contrib/minio`
- **Span name**: `minio-client.<HTTP method>` (e.g. `minio-client.PUT`)
- **Span Kind**: `SpanKindClient`
- Automatically injects trace context into request headers (W3C Trace Context propagation)

#### Slow Log

When request duration exceeds `SlowThreshold` (default 1000ms) or an error occurs, a slow log is automatically recorded:

```
[minio] PUT /my-bucket/key.txt bucket=my-bucket duration=2500ms err=<nil>
```

Set `SlowThreshold: 0` to disable slow log.

#### Circuit Breaker

- Independent circuit breaker per endpoint, named `minio:<endpoint>`
- 5xx responses and network errors are treated as failures
- When tripped, returns `breaker.ErrServiceUnavailable` and increments `minio_client_breaker_trip_total` metric

### Usage with go-zero

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
  # The following can be omitted, using default values
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
    // Upload file
    info, err := l.svcCtx.Minio.UploadFile(l.ctx, "my-bucket", "uploads/test.png", filePath,
        minio.WithContentType("image/png"),
    )
    if err != nil {
        return nil, err
    }

    // Generate presigned download URL (browser preview)
    url, err := l.svcCtx.Minio.GetPresignedDownloadURL(l.ctx, "my-bucket", "uploads/test.png",
        10*time.Minute,
        minio.WithResponseContentDisposition("inline"),
    )
    if err != nil {
        return nil, err
    }
    _ = url // return to frontend

    return info, nil
}
```

### Standalone Script Usage

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

    // Upload file
    info, err := client.UploadFile(ctx, "my-bucket", "test.txt", "/tmp/test.txt")
    if err != nil {
        fmt.Fprintf(os.Stderr, "upload failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Printf("uploaded: key=%s etag=%s size=%d\n", info.Key, info.ETag, info.Size)

    // Check if object exists
    exists, err := client.Exists(ctx, "my-bucket", "test.txt")
    if err != nil {
        fmt.Fprintf(os.Stderr, "exists check failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("exists:", exists)

    // Download
    reader, err := client.Download(ctx, "my-bucket", "test.txt")
    if err != nil {
        fmt.Fprintf(os.Stderr, "download failed: %v\n", err)
        os.Exit(1)
    }
    defer reader.Close()

    // Generate presigned URL
    url, err := client.GetPresignedDownloadURL(ctx, "my-bucket", "test.txt", 1*time.Hour)
    if err != nil {
        fmt.Fprintf(os.Stderr, "presign failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("presigned URL:", url)

    // Delete
    if err := client.Delete(ctx, "my-bucket", "test.txt"); err != nil {
        fmt.Fprintf(os.Stderr, "delete failed: %v\n", err)
        os.Exit(1)
    }
    fmt.Println("deleted")
}
```

## API Usage Examples

The following examples assume `client` has been created via `minio.MustNewClient(conf)` and `ctx` is a valid `context.Context`.

### Convenience Methods

#### UploadFile

```go
// Upload a local file with Content-Type and custom metadata
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

// Upload from a Reader (e.g. in-memory data, HTTP request body)
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
// Download an object; caller must close the reader
reader, err := client.Download(ctx, "my-bucket", "notes/hello.txt")
if err != nil {
    log.Fatal(err)
}
defer reader.Close()

io.Copy(os.Stdout, reader)
```

#### Delete

```go
// Delete an object
if err := client.Delete(ctx, "my-bucket", "notes/hello.txt"); err != nil {
    log.Fatal(err)
}
```

#### Exists

```go
// Check if an object exists
exists, err := client.Exists(ctx, "my-bucket", "notes/hello.txt")
if err != nil {
    log.Fatal(err)
}
fmt.Println("exists:", exists)
```

#### GetPresignedDownloadURL

```go
import "time"

// Generate a presigned download URL (preview image in browser)
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

// Generate a presigned upload URL; frontend can PUT directly
uploadURL, err := client.GetPresignedUploadURL(ctx, "my-bucket", "uploads/avatar.png", 15*time.Minute)
if err != nil {
    log.Fatal(err)
}
fmt.Println("upload URL:", uploadURL)
// Frontend usage: fetch(uploadURL, { method: 'PUT', body: file })
```

### Low-Level Operations

#### PutObject

```go
import (
    "strings"

    miniogo "github.com/minio/minio-go/v7"
)

// Full control over upload parameters
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

// Full control over download parameters (e.g. range download)
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

// Get object metadata (size, Content-Type, last modified, etc.)
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

// Full control over delete parameters (e.g. specify version ID)
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

// Copy an object to another location
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

// List all objects with a given prefix
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

### Bucket Management

#### MakeBucket

```go
import miniogo "github.com/minio/minio-go/v7"

// Create a bucket
err := client.MakeBucket(ctx, "new-bucket", miniogo.MakeBucketOptions{
    Region: "us-east-1",
})
if err != nil {
    log.Fatal(err)
}
```

#### RemoveBucket

```go
// Remove an empty bucket (ensure the bucket has no objects)
err := client.RemoveBucket(ctx, "empty-bucket")
if err != nil {
    log.Fatal(err)
}
```

#### ListBuckets

```go
// List all buckets
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
// Set bucket to public read-only policy
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
// Get current bucket policy
policy, err := client.GetBucketPolicy(ctx, "my-bucket")
if err != nil {
    log.Fatal(err)
}
fmt.Println("policy:", policy)
```

### Raw Client

#### RawClient

```go
import miniogo "github.com/minio/minio-go/v7"

// Get the P2C-selected underlying minio-go client for operations not wrapped by the SDK
raw := client.RawClient()

// Example: Use raw client to set bucket lifecycle
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
// Get all underlying clients for custom health checks or batch operations
clients := client.RawClients()
for i, raw := range clients {
    alive, _ := raw.HealthCheck(10 * time.Second)
    fmt.Printf("node[%d] healthy=%v\n", i, alive)
}
```

### WithTransport Example

```go
import (
    "crypto/tls"
    "net"
    "net/http"
    "time"

    "github.com/lerity-yao/czt-contrib/minio"
)

// Inject a custom Transport (e.g. custom timeout, TLS configuration)
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

### WithPartSize Example

```go
// Set part size to 64 MiB for large file uploads
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
