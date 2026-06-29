# czt-contrib

> Microservice governance component library for the go-zero ecosystem, production-proven since 2023

[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![GitHub tag](https://img.shields.io/github/tag/lerity-yao/czt-contrib.svg)](https://github.com/lerity-yao/czt-contrib/tags)
[![GitHub stars](https://img.shields.io/github/stars/lerity-yao/czt-contrib.svg)](https://github.com/lerity-yao/czt-contrib/stargazers)
[![GitHub last commit](https://img.shields.io/github/last-commit/lerity-yao/czt-contrib.svg)](https://github.com/lerity-yao/czt-contrib/commits)
[![Go version](https://img.shields.io/github/go-mod/go-version/lerity-yao/czt-contrib.svg)](https://github.com/lerity-yao/czt-contrib/blob/main/go.mod)
[![codecov](https://codecov.io/gh/lerity-yao/czt-contrib/branch/main/graph/badge.svg)](https://codecov.io/gh/lerity-yao/czt-contrib)

[中文](./readme-cn.md)

## Background & Motivation

In 2023, the team underwent a thorough architecture overhaul. Faced with rapidly growing business traffic and increasingly complex service topology, we chose go-zero as the microservices foundation—its high performance, low latency, and built-in governance capabilities solidly solved core challenges such as RPC, API gateways, and circuit breaking. However, as the business evolved, we found that the ecosystem around go-zero still needed to be filled in ourselves: service registry, configuration center, message queue, scheduled tasks, distributed ID, Alibaba Cloud gateway integration, Kong HMAC signing… Almost every new project repeated the wheel, with component implementations varying in style and monitoring, tracing, and error governance lacking unified standards. It was against this backdrop that czt-contrib emerged. It is not meant to replace go-zero, but to provide a production-hardened unified layer around go-zero, enabling teams to gain standardized, observable, and governable microservice capabilities with minimal integration cost. Since 2023, these components have been running stably in internal core businesses, withstanding both major promotion traffic and daily iterations.

## Design Philosophy

- **Native go-zero integration**: Configuration, circuit breaker, and tracing seamlessly connect, fitting into existing engineering standards upon adoption.
- **Standardized observability**: Unified integration with OpenTelemetry and Prometheus, not tied to any specific cloud platform, and pluggable into any observability backend.
- **Plug and play**: Each sub-module is released independently and ready to use out of the box, avoiding extra glue code and repeated encapsulation.
- **Generality first**: No dependency on go-zero private APIs; even non-go-zero projects can use them independently as needed.

## Module Overview

| Module | Description | Docs |
|------|------|------|
| cztctl | Code generation tool for the component library | [README](./cztctl/README.md) |
| configcenter/consul | Distributed configuration center based on Consul | [README](./configcenter/consul/README.md) |
| registercenter/consul | Service registration and discovery based on Consul | [README](./registercenter/consul/README.md) |
| snake | High-concurrency Snowflake distributed ID generator | [README](./snake/README.md) |
| mq/rabbitmq | RabbitMQ message queue client | [README](./mq/rabbitmq/README.md) |
| cron | Distributed task queue based on Redis | [README](./cron/README.md) |
| aliyun/gateway | Alibaba Cloud API Gateway Go client | [README](./aliyun/gateway/README.md) |
| kong/hmacauth | Kong HMAC Auth signing client | [README](./kong/hmacauth/README.md) |
| minio | MinIO object storage client with P2C load balancing and go-zero integration | [README](./minio/README.md) |

## Sub-module Quality Report

| Module | Go Report Card | Go Reference | goproxy.cn | Codecov |
|------|-----------------|--------------|------------|---------|
| [aliyun/gateway](./aliyun/gateway) | [![Go Report Card](https://goreportcard.com/badge/github.com/lerity-yao/czt-contrib/aliyun/gateway)](https://goreportcard.com/report/github.com/lerity-yao/czt-contrib/aliyun/gateway) | [![Go Reference](https://pkg.go.dev/badge/github.com/lerity-yao/czt-contrib/aliyun/gateway.svg)](https://pkg.go.dev/github.com/lerity-yao/czt-contrib/aliyun/gateway) | [![goproxy](https://goproxy.cn/stats/github.com/lerity-yao/czt-contrib/aliyun/gateway/badges/download-count.svg)](https://goproxy.cn/stats/github.com/lerity-yao/czt-contrib/aliyun/gateway/badges/download-count.svg) | [![codecov](https://codecov.io/gh/lerity-yao/czt-contrib/branch/main/graph/badge.svg?flag=aliyun-gateway)](https://codecov.io/gh/lerity-yao/czt-contrib) |
| [aliyun/oss](./aliyun/oss) | [![Go Report Card](https://goreportcard.com/badge/github.com/lerity-yao/czt-contrib/aliyun/oss)](https://goreportcard.com/report/github.com/lerity-yao/czt-contrib/aliyun/oss) | [![Go Reference](https://pkg.go.dev/badge/github.com/lerity-yao/czt-contrib/aliyun/oss.svg)](https://pkg.go.dev/github.com/lerity-yao/czt-contrib/aliyun/oss) | [![goproxy](https://goproxy.cn/stats/github.com/lerity-yao/czt-contrib/aliyun/oss/badges/download-count.svg)](https://goproxy.cn/stats/github.com/lerity-yao/czt-contrib/aliyun/oss/badges/download-count.svg) | |
| [configcenter/consul](./configcenter/consul) | [![Go Report Card](https://goreportcard.com/badge/github.com/lerity-yao/czt-contrib/configcenter/consul)](https://goreportcard.com/report/github.com/lerity-yao/czt-contrib/configcenter/consul) | [![Go Reference](https://pkg.go.dev/badge/github.com/lerity-yao/czt-contrib/configcenter/consul.svg)](https://pkg.go.dev/github.com/lerity-yao/czt-contrib/configcenter/consul) | [![goproxy](https://goproxy.cn/stats/github.com/lerity-yao/czt-contrib/configcenter/consul/badges/download-count.svg)](https://goproxy.cn/stats/github.com/lerity-yao/czt-contrib/configcenter/consul/badges/download-count.svg) | [![codecov](https://codecov.io/gh/lerity-yao/czt-contrib/branch/main/graph/badge.svg?flag=configcenter-consul)](https://codecov.io/gh/lerity-yao/czt-contrib) |
| [registercenter/consul](./registercenter/consul) | [![Go Report Card](https://goreportcard.com/badge/github.com/lerity-yao/czt-contrib/registercenter/consul)](https://goreportcard.com/report/github.com/lerity-yao/czt-contrib/registercenter/consul) | [![Go Reference](https://pkg.go.dev/badge/github.com/lerity-yao/czt-contrib/registercenter/consul.svg)](https://pkg.go.dev/github.com/lerity-yao/czt-contrib/registercenter/consul) | [![goproxy](https://goproxy.cn/stats/github.com/lerity-yao/czt-contrib/registercenter/consul/badges/download-count.svg)](https://goproxy.cn/stats/github.com/lerity-yao/czt-contrib/registercenter/consul/badges/download-count.svg) | [![codecov](https://codecov.io/gh/lerity-yao/czt-contrib/branch/main/graph/badge.svg?flag=registercenter-consul)](https://codecov.io/gh/lerity-yao/czt-contrib) |
| [snake](./snake) | [![Go Report Card](https://goreportcard.com/badge/github.com/lerity-yao/czt-contrib/snake)](https://goreportcard.com/report/github.com/lerity-yao/czt-contrib/snake) | [![Go Reference](https://pkg.go.dev/badge/github.com/lerity-yao/czt-contrib/snake.svg)](https://pkg.go.dev/github.com/lerity-yao/czt-contrib/snake) | [![goproxy](https://goproxy.cn/stats/github.com/lerity-yao/czt-contrib/snake/badges/download-count.svg)](https://goproxy.cn/stats/github.com/lerity-yao/czt-contrib/snake/badges/download-count.svg) | [![codecov](https://codecov.io/gh/lerity-yao/czt-contrib/branch/main/graph/badge.svg?flag=snake)](https://codecov.io/gh/lerity-yao/czt-contrib) |
| [mq/rabbitmq](./mq/rabbitmq) | [![Go Report Card](https://goreportcard.com/badge/github.com/lerity-yao/czt-contrib/mq/rabbitmq)](https://goreportcard.com/report/github.com/lerity-yao/czt-contrib/mq/rabbitmq) | [![Go Reference](https://pkg.go.dev/badge/github.com/lerity-yao/czt-contrib/mq/rabbitmq.svg)](https://pkg.go.dev/github.com/lerity-yao/czt-contrib/mq/rabbitmq) | [![goproxy](https://goproxy.cn/stats/github.com/lerity-yao/czt-contrib/mq/rabbitmq/badges/download-count.svg)](https://goproxy.cn/stats/github.com/lerity-yao/czt-contrib/mq/rabbitmq/badges/download-count.svg) | |
| [cron](./cron) | [![Go Report Card](https://goreportcard.com/badge/github.com/lerity-yao/czt-contrib/cron)](https://goreportcard.com/report/github.com/lerity-yao/czt-contrib/cron) | [![Go Reference](https://pkg.go.dev/badge/github.com/lerity-yao/czt-contrib/cron.svg)](https://pkg.go.dev/github.com/lerity-yao/czt-contrib/cron) | [![goproxy](https://goproxy.cn/stats/github.com/lerity-yao/czt-contrib/cron/badges/download-count.svg)](https://goproxy.cn/stats/github.com/lerity-yao/czt-contrib/cron/badges/download-count.svg) | [![codecov](https://codecov.io/gh/lerity-yao/czt-contrib/branch/main/graph/badge.svg?flag=cron)](https://codecov.io/gh/lerity-yao/czt-contrib) |
| [cztctl](./cztctl) | [![Go Report Card](https://goreportcard.com/badge/github.com/lerity-yao/czt-contrib/cztctl)](https://goreportcard.com/report/github.com/lerity-yao/czt-contrib/cztctl) | [![Go Reference](https://pkg.go.dev/badge/github.com/lerity-yao/czt-contrib/cztctl.svg)](https://pkg.go.dev/github.com/lerity-yao/czt-contrib/cztctl) | [![goproxy](https://goproxy.cn/stats/github.com/lerity-yao/czt-contrib/cztctl/badges/download-count.svg)](https://goproxy.cn/stats/github.com/lerity-yao/czt-contrib/cztctl/badges/download-count.svg) | |
| [kong/hmacauth](./kong/hmacauth) | [![Go Report Card](https://goreportcard.com/badge/github.com/lerity-yao/czt-contrib/kong/hmacauth)](https://goreportcard.com/report/github.com/lerity-yao/czt-contrib/kong/hmacauth) | [![Go Reference](https://pkg.go.dev/badge/github.com/lerity-yao/czt-contrib/kong/hmacauth.svg)](https://pkg.go.dev/github.com/lerity-yao/czt-contrib/kong/hmacauth) | [![goproxy](https://goproxy.cn/stats/github.com/lerity-yao/czt-contrib/kong/hmacauth/badges/download-count.svg)](https://goproxy.cn/stats/github.com/lerity-yao/czt-contrib/kong/hmacauth/badges/download-count.svg) | [![codecov](https://codecov.io/gh/lerity-yao/czt-contrib/branch/main/graph/badge.svg?flag=kong-hmacauth)](https://codecov.io/gh/lerity-yao/czt-contrib) |
| [minio](./minio) | [![Go Report Card](https://goreportcard.com/badge/github.com/lerity-yao/czt-contrib/minio)](https://goreportcard.com/report/github.com/lerity-yao/czt-contrib/minio) | [![Go Reference](https://pkg.go.dev/badge/github.com/lerity-yao/czt-contrib/minio.svg)](https://pkg.go.dev/github.com/lerity-yao/czt-contrib/minio) | [![goproxy](https://goproxy.cn/stats/github.com/lerity-yao/czt-contrib/minio/badges/download-count.svg)](https://goproxy.cn/stats/github.com/lerity-yao/czt-contrib/minio/badges/download-count.svg) | [![codecov](https://codecov.io/gh/lerity-yao/czt-contrib/branch/main/graph/badge.svg?flag=minio)](https://codecov.io/gh/lerity-yao/czt-contrib) |

## Quick Start

Taking `snake` as an example, you can generate a distributed unique ID in just two steps:

```bash
go get github.com/lerity-yao/czt-contrib/snake
```

```go
package main

import (
    "fmt"

    "github.com/lerity-yao/czt-contrib/snake"
)

func main() {
    s := snake.MustNewSnake(snake.Conf{
        WorkerIDBits:   10,
        SequenceBits:   12,
        WorkerID:       1,
    })

    id, _ := s.Generator()
    fmt.Printf("Generated ID: %d\n", id)
}
```

## Project Structure

```
czt-contrib/
├── cztctl/           # Code generation tool
├── configcenter/      # Configuration center module
│   ├── consul/       # Consul configuration center implementation
│   └── nacos/        # Nacos configuration center implementation (to be implemented)
├── registercenter/   # Service registry module
│   └── consul/       # Consul service registry implementation
├── snake/            # Distributed ID generator module
├── mq/               # Message queue module
│   └── rabbitmq/     # RabbitMQ implementation
├── cron/             # Distributed task queue module
├── aliyun/           # Alibaba Cloud component module
│   └── gateway/     # API Gateway client
├── kong/             # Kong component module
│   └── hmacauth/    # Kong HMAC Auth client
├── go.mod            # Go module definition
└── go.sum            # Go dependency checksum
```

## Modules

### Code Generation Tool (cztctl)

A code generation tool extended based on goctl, designed specifically for the czt-contrib component library. It supports one-click generation of RabbitMQ consumer services, Cron scheduled task services, Swagger documentation, and more.

**Core Capabilities**:
- `cztctl api swagger` - Generate Swagger API docs from `.api` files
- `cztctl api cron` - Generate distributed scheduled task service code from `.cron` files
- `cztctl api rabbitmq` - Generate RabbitMQ consumer service code from `.rabbitmq` files
- `cztctl rpc sdk` - Generate RPC client SDK and publish it to an independent Git repository
- `cztctl env` - View or edit cztctl environment variables
- Generated code automatically integrates tracing, metrics, error handling, and other capabilities

**IDE Plugins**:
- [cztctl-intellij](https://github.com/lerity-yao/cztctl-intellij) - GoLand/IntelliJ IDEA syntax highlighting plugin supporting `.cron` and `.rabbitmq` files with syntax highlighting, code completion, and error checking
- [cztctl-vscode](https://github.com/lerity-yao/cztctl-vscode) - VS Code syntax highlighting plugin supporting `.cron` and `.rabbitmq` files with syntax highlighting, code completion, and error checking

For details, see [cztctl/README.md](./cztctl/README.md).

### Configuration Center (configcenter)

A distributed configuration management center based on HashiCorp Consul, designed for the go-zero framework.

**Core Capabilities**:
- Consul KV-based configuration subscriber with change detection via Long-Polling + Index
- Supports multiple configuration formats: YAML, HCL, JSON, XML, automatically normalized to JSON
- Event-driven updates: register callbacks via `AddListener` to be notified automatically on configuration changes
- Background goroutine continuously listens, with automatic retry on fetch failures (1-second interval)
- Supports TLS encrypted connections and Token authentication

For details, see [configcenter/consul/README.md](./configcenter/consul/README.md).

### Service Registry (registercenter)

A service registration and discovery center based on HashiCorp Consul, designed for the go-zero framework.

**Core Capabilities**:
- Automatic service registration and deregistration, with automatic deregistration on process exit
- Three health check mechanisms: TTL heartbeat, HTTP endpoint, and gRPC Health protocol
- Health monitoring and automatic recovery: exponential backoff retry on failures (up to 5 times, 1s→30s), and automatic re-registration when health loss is detected
- gRPC Resolver-based service discovery: automatically registers the `consul://` scheme, supporting query parameters such as tag/dc/near
- Container environment adaptation: automatically recognizes the `POD_IP` environment variable, and resolves `0.0.0.0` to the actual IP

For details, see [registercenter/consul/README.md](./registercenter/consul/README.md).

### Distributed ID Generator (snake)

A distributed unique ID generator based on the Snowflake algorithm, suitable for high-concurrency scenarios.

**Core Capabilities**:
- Based on a Snowflake algorithm variant, generates globally unique, ordered 64-bit IDs
- Lock-free design: uses atomic CAS for concurrency safety without locks
- Configurable bit allocation: WorkerIDBits (default 10) + SequenceBits (default 12), validated to sum ≤ 63
- Automatic worker node ID allocation: based on FNV hash of the IP address, with manual specification also supported
- Clock rollback handling: configurable tolerance (default 5ms), rejects generation if exceeded
- ID parsing: timestamp, worker node ID, and sequence number can be extracted from the ID

For details, see [snake/README.md](./snake/README.md).

### Message Queue (mq/rabbitmq)

A high-performance message queue client built on [RabbitMQ](https://www.rabbitmq.com/), a distributed message processing module designed for the go-zero framework.

**Core Capabilities**:
- Sender (producer): single-connection persistent heartbeat (30s), automatic exponential backoff reconnection on disconnect (max 30s)
- Listener (consumer): implements the go-zero `queue.MessageQueue` interface, supports QoS prefetch control, manual/auto ACK
- Integrated OpenTelemetry tracing: Trace Context propagation across services via HeaderCarrier
- Built-in Prometheus metrics: Sender 5 + Listener 9 (send/consume count, latency, bytes, reconnections, disconnections, in-flight messages, etc.)
- Interceptor chain: Recovery (panic capture) → Trace → Prometheus → Logging
- Graceful shutdown: coordinated via go-zero proc, waiting for in-flight messages to be processed

For details, see [mq/rabbitmq/README.md](./mq/rabbitmq/README.md).

### Distributed Task Queue (cron)

A distributed task queue system built on [Asynq](https://github.com/hibiken/asynq), a scheduled and asynchronous task processing module designed for the go-zero framework.

**Core Capabilities**:
- Distributed task queue based on [Asynq](https://github.com/hibiken/asynq)
- Client (producer): supports immediate/delayed/scheduled push, automatic JSON serialization, automatic Trace Context injection
- Server (consumer): supports Cron expression scheduled tasks, concurrency control (default CPU cores), priority queues
- Supports multiple Redis modes: standalone, sentinel, cluster, with TLS support
- Middleware chain: Recovery (panic capture without retry) → Prometheus → Trace
- Built-in Prometheus metrics: consume count, latency, bytes, active workers, retry count, panic count, etc.
- Task management: supports task cancellation (CancelTask) and rescheduling (RescheduleTask)

For details, see [cron/README.md](./cron/README.md).

### Alibaba Cloud API Gateway Client (aliyun/gateway)

An Alibaba Cloud API Gateway Go client built on go-zero httpc, automatically completing HMAC-SHA256 v1 signing.

**Core Capabilities**:
- Automatically completes Alibaba Cloud API Gateway v1 HMAC-SHA256 signing, injecting the full set of `X-Ca-*` headers (Key, Nonce, Timestamp, Signature-Method, Signature-Headers, Signature)
- `Do` method: structured requests, automatically mapping path/form/json/header tags
- `DoRaw` method: raw byte requests, suitable for file uploads, XML, plain text, and other scenarios
- Automatically calculates Content-MD5 (except form/multipart)
- Underlying integration with go-zero httpc, automatically sharing circuit breakers for the same Host
- Supports injecting a custom `*http.Client` via `WithClient` (timeout, TLS, connection pool)

For details, see [aliyun/gateway/README.md](./aliyun/gateway/README.md).

### Kong HMAC Auth Client (kong/hmacauth)

A Kong HMAC Auth Go client built on go-zero httpc, automatically completing HMAC signing and following the official [Kong HMAC Auth plugin](https://developer.konghq.com/plugins/hmac-auth/) specification.

**Core Capabilities**:
- Supports 5 HMAC algorithms: hmac-sha1 / sha224 / sha256 / sha384 / sha512
- `@request-target` pseudo-header signing, compliant with the Kong 3.x official specification
- Automatically injects Date (RFC 2822 UTC), User-Agent, Host, and Digest headers
- Digest integrity: includes SHA-256 digest calculation for empty-body scenarios (except form/multipart)
- `Do`/`DoRaw` dual methods: structured requests + raw byte requests
- Customizable list of headers participating in signing, default `["date", "@request-target"]`
- Underlying integration with go-zero httpc, automatically sharing circuit breakers for the same Host
- Supports injecting a custom `*http.Client` via `WithClient`

For details, see [kong/hmacauth/README.md](./kong/hmacauth/README.md).

## Contributors

Thanks to all the developers who have contributed to this project:

<a href="https://github.com/lerity-yao/czt-contrib/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=lerity-yao/czt-contrib" />
</a>

## Contributing & License

Contributions via Issues and Pull Requests are welcome. The project is licensed under MIT. For details, see [LICENSE](LICENSE).
