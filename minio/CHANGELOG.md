# Changelog

[中文](./changelog-cn.md)

All version change logs. Format based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [0.0.1] - 2026-06-29

### Added

- MinIO Go client SDK built on minio-go v7, integrated with the go-zero ecosystem, providing a unified object storage operation interface
- `Client` interface: defines four API layers — convenience methods, low-level operations, bucket management, and raw client access
- `NewClient` / `MustNewClient` constructors: support `ClientOption` optional parameter injection
- `WithTransport` option: inject a custom base `http.RoundTripper`; the instrumented Transport wraps it automatically

#### Convenience Methods

- `UploadFile`: upload a local file with automatic multipart and Option parameters (ContentType, Metadata, StorageClass, PartSize)
- `UploadReader`: upload from an `io.Reader` with the same Option parameters
- `Download`: download an object, returns an `io.ReadCloser` stream with affinity-aware node selection
- `Delete`: delete an object with P2C failover support
- `Exists`: check whether an object exists, automatically handles `NoSuchKey` / 404 responses
- `GetPresignedDownloadURL`: generate a presigned download URL with `PresignedOption` support (`WithResponseContentDisposition`, `WithResponseContentType`)
- `GetPresignedUploadURL`: generate a presigned upload URL

#### Low-Level Operations

- `PutObject`: upload with full control over `miniogo.PutObjectOptions`
- `GetObject`: download with full control over `miniogo.GetObjectOptions`, affinity-aware selection
- `StatObject`: get object metadata with affinity-aware + failover
- `RemoveObject`: delete with full control over `miniogo.RemoveObjectOptions`
- `CopyObject`: copy an object, sets affinity on write
- `ListObjects`: list objects, returns a channel, P2C selection

#### Bucket Management

- `MakeBucket`: create a bucket
- `RemoveBucket`: remove an empty bucket
- `ListBuckets`: list all buckets
- `SetBucketPolicy` / `GetBucketPolicy`: set/get bucket policy

#### Raw Client

- `RawClient`: get the P2C-selected underlying minio-go client
- `RawClients`: get all underlying minio-go clients

#### P2C Load Balancing

- Power of Two Choices load balancing algorithm implementation based on EWMA latency × inflight load scoring
- EWMA decay window of 10 seconds, automatically smoothing latency jitter
- `forcePick` mechanism: nodes not selected for more than 1 second are force-selected to avoid starvation
- Supports three modes: single-node direct selection, two-node direct comparison, multi-node random-two comparison

#### Write-After-Read Affinity

- Write-after-read affinity cache based on go-zero `collection.Cache` (TimingWheel-driven)
- After a successful write, automatically records `bucket/key → nodeIndex` mapping with 5-second TTL
- Read operations first query the affinity cache; on hit, route to the write node
- On affinity read failure, penalizes the node's load (+penalty) and falls back to P2C failover
- Affinity cache naming isolated by AccessKeyID: `minio:affinity:<ak>`

#### Failover

- All non-streaming operations support automatic failover: shuffle remaining nodes with deterministic traversal after preferred node network failure
- Network error detection: `net.Error` type matching + keyword fallback (connection refused, connection reset, no such host, i/o timeout, network is unreachable, dial tcp)
- Non-network errors (e.g. 4xx business errors) do not trigger failover

#### Observability

- `instrumentedTransport`: wraps `http.RoundTripper`, integrating tracing, circuit breaker, metrics, and slow log
- OpenTelemetry tracing: creates a `minio-client.<METHOD>` Span for each HTTP request, automatically injects trace context
- go-zero circuit breaker: independent breaker per endpoint (`minio:<endpoint>`), 5xx and network errors treated as failures
- Prometheus metrics: `minio_client_requests_duration_ms` (Histogram), `minio_client_requests_code_total` (Counter), `minio_client_requests_failover_total` (Counter), `minio_client_affinity_hit_total` (Counter), `minio_client_affinity_miss_total` (Counter), `minio_client_breaker_trip_total` (Counter)
- Slow log: automatically recorded when request duration exceeds `SlowThreshold` or an error occurs; `SlowThreshold: 0` disables it

#### Type System

- `UploadInfo`: upload result (Key, ETag, Size, VersionID)
- `ObjectInfo`: object metadata (Key, Size, ContentType, ETag, LastModified, Metadata)
- `BucketInfo`: bucket information (Name, CreationDate)
- `Error`: custom error type wrapping minio-go `ErrorResponse`, providing Code, Message, BucketName, Key, StatusCode, Cause structured fields, supporting `errors.As` and `errors.Unwrap`

#### Configuration

- `Conf` struct: supports `Endpoints`, `AccessKeyID`, `SecretAccessKey`, `UseSSL`, `Region`, `SignatureVersion`, `SlowThreshold` configuration
- `Validate()` method: validates required fields, non-empty endpoints, valid signature version, non-negative slow log threshold
- V2/V4 dual signature support: select `v2` or `v4` signature algorithm via the `SignatureVersion` field

