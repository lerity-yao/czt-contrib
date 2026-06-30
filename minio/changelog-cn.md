# Changelog

[English](./CHANGELOG.md)

所有版本变更记录。格式参考 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)。

## [0.0.1] - 2026-06-29

### Added

- 基于 minio-go v7 封装的 MinIO Go 客户端 SDK，集成 go-zero 生态，提供统一的对象存储操作接口
- `Client` 接口：定义便捷方法、原子操作、Bucket 管理、原始客户端访问四层 API
- `NewClient` / `MustNewClient` 构造函数：支持 `ClientOption` 可选参数注入
- `WithTransport` 选项：注入自定义 base `http.RoundTripper`，可观测 Transport 自动包装

#### 便捷方法

- `UploadFile`：上传本地文件，支持自动分片和 Option 参数（ContentType、Metadata、StorageClass、PartSize）
- `UploadReader`：从 `io.Reader` 上传，支持同上 Option 参数
- `Download`：下载对象，返回 `io.ReadCloser` 流，使用亲和感知节点选择
- `Delete`：删除对象，支持 P2C 故障转移
- `Exists`：判断对象是否存在，自动处理 `NoSuchKey` / 404 响应
- `GetPresignedDownloadURL`：生成预签名下载 URL，支持 `PresignedOption`（`WithResponseContentDisposition`、`WithResponseContentType`）
- `GetPresignedUploadURL`：生成预签名上传 URL

#### 原子操作

- `PutObject`：完全控制 `miniogo.PutObjectOptions` 的上传操作
- `GetObject`：完全控制 `miniogo.GetObjectOptions` 的下载操作，亲和感知选择
- `StatObject`：获取对象元数据，亲和感知 + 故障转移
- `RemoveObject`：完全控制 `miniogo.RemoveObjectOptions` 的删除操作
- `CopyObject`：复制对象，写操作设置亲和
- `ListObjects`：列举对象，返回 channel，P2C 选择

#### Bucket 管理

- `MakeBucket`：创建 Bucket
- `RemoveBucket`：删除空 Bucket
- `ListBuckets`：列举所有 Bucket
- `SetBucketPolicy` / `GetBucketPolicy`：设置/获取 Bucket 策略

#### 原始客户端

- `RawClient`：获取 P2C 选中的底层 minio-go 客户端
- `RawClients`：获取全部底层 minio-go 客户端

#### P2C 负载均衡

- Power of Two Choices 负载均衡算法实现，基于 EWMA 延迟 × 并发数的负载评分
- EWMA 衰减窗口 10 秒，自动平滑延迟抖动
- `forcePick` 机制：超过 1 秒未被选中的节点强制选择，避免饥饿
- 支持单节点直选、双节点直接比较、多节点随机取二比较三种模式

#### 写后读亲和

- 基于 go-zero `collection.Cache`（TimingWheel 驱动）的写后读亲和缓存
- 写操作成功后自动记录 `bucket/key → nodeIndex` 映射，TTL 5 秒
- 读操作优先查询亲和缓存，命中时路由到写入节点
- 亲和读失败时惩罚该节点负载（+penalty）并回退到 P2C 故障转移
- 亲和缓存命名基于 AccessKeyID 隔离：`minio:affinity:<ak>`

#### 故障转移

- 所有非流式操作支持自动故障转移：首选节点网络失败后 shuffle 剩余节点确定性遍历
- 网络错误检测：`net.Error` 类型匹配 + 关键词兜底（connection refused、connection reset、no such host、i/o timeout、network is unreachable、dial tcp）
- 非网络错误（如 4xx 业务错误）不触发故障转移

#### 可观测性

- `instrumentedTransport`：包装 `http.RoundTripper`，集成追踪、熔断、指标、慢日志
- OpenTelemetry 追踪：每个 HTTP 请求创建 `minio-client.<METHOD>` Span，自动注入 trace context
- go-zero 熔断器：每个 endpoint 独立熔断（`minio:<endpoint>`），5xx 和网络错误视为失败
- Prometheus 指标：`minio_client_requests_duration_ms`（Histogram）、`minio_client_requests_code_total`（Counter）、`minio_client_requests_failover_total`（Counter）、`minio_client_affinity_hit_total`（Counter）、`minio_client_affinity_miss_total`（Counter）、`minio_client_breaker_trip_total`（Counter）
- 慢日志：请求耗时超过 `SlowThreshold` 或发生错误时自动记录，`SlowThreshold: 0` 禁用

#### 类型系统

- `UploadInfo`：上传结果（Key、ETag、Size、VersionID）
- `ObjectInfo`：对象元数据（Key、Size、ContentType、ETag、LastModified、Metadata）
- `BucketInfo`：Bucket 信息（Name、CreationDate）
- `Error`：自定义错误类型，包装 minio-go `ErrorResponse`，提供 Code、Message、BucketName、Key、StatusCode、Cause 结构化字段，支持 `errors.As` 和 `errors.Unwrap`

#### 配置

- `Conf` 结构体：支持 `Endpoints`、`AccessKeyID`、`SecretAccessKey`、`UseSSL`、`Region`、`SignatureVersion`、`SlowThreshold` 配置
- `Validate()` 校验方法：检查必传字段、endpoint 非空、签名版本有效、慢日志阈值非负
- V2/V4 双签名支持：通过 `SignatureVersion` 字段选择 `v2` 或 `v4` 签名算法

