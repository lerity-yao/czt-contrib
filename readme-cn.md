# czt-contrib

> 面向 go-zero 生态的微服务治理组件库，自 2023 年起生产验证

[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)
[![GitHub tag](https://img.shields.io/github/tag/lerity-yao/czt-contrib.svg)](https://github.com/lerity-yao/czt-contrib/tags)
[![GitHub stars](https://img.shields.io/github/stars/lerity-yao/czt-contrib.svg)](https://github.com/lerity-yao/czt-contrib/stargazers)
[![GitHub last commit](https://img.shields.io/github/last-commit/lerity-yao/czt-contrib.svg)](https://github.com/lerity-yao/czt-contrib/commits)
[![Go version](https://img.shields.io/github/go-mod/go-version/lerity-yao/czt-contrib.svg)](https://github.com/lerity-yao/czt-contrib/blob/main/go.mod)
[![codecov](https://codecov.io/gh/lerity-yao/czt-contrib/branch/main/graph/badge.svg)](https://codecov.io/gh/lerity-yao/czt-contrib)

[English](./README.md)

## 背景与动机

2023 年，团队经历了一次彻底的架构调整。面对快速增长的业务流量和日益复杂的服务拓扑，我们选择以 go-zero 作为微服务底座——它的高性能、低延迟和内置的治理能力的确实打实地解决了 RPC、API 网关和限流熔断等核心问题。但随着业务深入，我们发现 go-zero 周边的生态拼图仍需自己补齐：注册中心、配置中心、消息队列、定时任务、分布式 ID、阿里云网关对接、Kong HMAC 签名……每一个新项目几乎都在重复造轮子，组件实现风格各异，监控、链路追踪和错误治理也缺乏统一标准。正是在这样的背景下，czt-contrib 应运而生。它并不是要替代 go-zero，而是围绕 go-zero 做一层经过生产打磨的统一封装，让团队能够以最小的接入成本获得标准化、可观测、可治理的微服务能力。自 2023 年至今，这些组件已在内部核心业务中长期稳定运行，经历了大促流量与日常迭代的双重考验。

## 设计理念

- **go-zero 原生集成**：配置体系、熔断器、链路追踪无缝对接，接入即融入既有工程规范。
- **标准化可观测**：统一集成 OpenTelemetry 与 Prometheus，不绑定特定云平台，可插拔接入任意观测后端。
- **即插即用**：每个子模块独立发版，引入即可用，避免额外的胶水代码与重复封装。
- **通用性优先**：不依赖 go-zero 私有 API，即使非 go-zero 项目也能按需独立使用。

## 模块概览

| 模块 | 描述 | 文档 |
|------|------|------|
| cztctl | 面向组件库的代码生成工具 | [README](./cztctl/README.md) |
| configcenter/consul | 基于 Consul 的分布式配置中心 | [README](./configcenter/consul/README.md) |
| registercenter/consul | 基于 Consul 的服务注册与发现 | [README](./registercenter/consul/README.md) |
| snake | 高并发雪花算法分布式 ID 生成器 | [README](./snake/README.md) |
| mq/rabbitmq | RabbitMQ 消息队列客户端 | [README](./mq/rabbitmq/README.md) |
| cron | 基于 Redis 的分布式任务队列 | [README](./cron/README.md) |
| aliyun/gateway | 阿里云 API 网关 Go 客户端 | [README](./aliyun/gateway/README.md) |
| kong/hmacauth | Kong HMAC Auth 签名客户端 | [README](./kong/hmacauth/README.md) |
| minio | MinIO 对象存储客户端，集成 P2C 负载均衡、写后读亲和、熔断器和可观测性 | [README](./minio/README.md) |

## 子模块质量报告

| 模块 | Go Report Card | Go Reference | goproxy.cn | Codecov |
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

## 快速开始

以 `snake` 为例，只需两步即可生成分布式唯一 ID：

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

## 项目结构

```
czt-contrib/
├── cztctl/           # 代码生成工具
├── configcenter/      # 配置中心模块
│   ├── consul/       # Consul 配置中心实现
│   └── nacos/        # Nacos 配置中心实现（待实现）
├── registercenter/   # 服务注册中心模块
│   └── consul/       # Consul 服务注册实现
├── snake/            # 分布式 ID 生成器模块
├── mq/               # 消息队列模块
│   └── rabbitmq/     # RabbitMQ 实现
├── cron/             # 分布式任务队列模块
├── aliyun/           # 阿里云组件模块
│   └── gateway/     # API 网关客户端
├── kong/             # Kong 组件模块
│   └── hmacauth/    # Kong HMAC Auth 客户端
├── go.mod            # Go 模块定义
└── go.sum            # Go 依赖校验
```

## 功能模块

### 代码生成工具 (cztctl)

基于 goctl 扩展的代码生成工具，专为 czt-contrib 组件库设计，支持一键生成 RabbitMQ 消费者服务、Cron 定时任务服务、Swagger 文档等。

**核心能力**：
- `cztctl api swagger` - 根据 `.api` 文件生成 Swagger API 文档
- `cztctl api cron` - 根据 `.cron` 文件生成分布式定时任务服务代码
- `cztctl api rabbitmq` - 根据 `.rabbitmq` 文件生成 RabbitMQ 消费者服务代码
- `cztctl rpc sdk` - 生成 RPC 客户端 SDK 并发布到独立 Git 仓库
- `cztctl env` - 查看或编辑 cztctl 环境变量
- 生成代码自动集成链路追踪、监控指标、错误处理等能力

**IDE 插件**：
- [cztctl-intellij](https://github.com/lerity-yao/cztctl-intellij) - GoLand/IntelliJ IDEA 语法高亮插件，支持 `.cron`、`.rabbitmq` 文件的语法高亮、代码补全、错误检查
- [cztctl-vscode](https://github.com/lerity-yao/cztctl-vscode) - VS Code 语法高亮插件，支持 `.cron`、`.rabbitmq` 文件的语法高亮、代码补全、错误检查

详情请参见：[cztctl/README.md](./cztctl/README.md)

### 配置中心 (configcenter)

基于 HashiCorp Consul 实现的分布式配置管理中心，专为 go-zero 框架设计。

**核心能力**：
- 基于 Consul KV 的配置订阅器，通过 Long-Polling + Index 实现变更检测
- 支持多种配置格式：YAML、HCL、JSON、XML，自动归一化为 JSON
- 事件驱动更新：通过 `AddListener` 注册回调，配置变更时自动通知
- 后台 Goroutine 持续监听，fetch 失败自动重试（1 秒间隔）
- 支持 TLS 加密连接和 Token 认证

详情请参见：[configcenter/consul/README.md](./configcenter/consul/README.md)

### 服务注册中心 (registercenter)

基于 HashiCorp Consul 实现的服务注册与发现中心，专为 go-zero 框架设计。

**核心能力**：
- 服务自动注册与注销，进程退出时自动反注册
- 三种健康检查机制：TTL 心跳、HTTP 端点、gRPC Health 协议
- 健康监控与自动恢复：失败时指数退避重试（最多 5 次，1s→30s），检测到健康丢失自动重新注册
- 基于 gRPC Resolver 的服务发现：自动注册 `consul://` scheme，支持 tag/dc/near 等查询参数
- 容器环境适配：自动识别 `POD_IP` 环境变量，`0.0.0.0` 自动解析为实际 IP

详情请参见：[registercenter/consul/README.md](./registercenter/consul/README.md)

### 分布式 ID 生成器 (snake)

基于雪花算法实现的分布式唯一 ID 生成器，适用于高并发场景。

**核心能力**：
- 基于雪花算法变体，生成全局唯一、有序的 64 位 ID
- 无锁设计：使用 atomic CAS 实现并发安全，无需加锁
- 可配置位分配：WorkerIDBits（默认 10）+ SequenceBits（默认 12），校验总和 ≤ 63
- 工作节点 ID 自动分配：基于 IP 地址的 FNV 哈希计算，也支持手动指定
- 时钟回拨处理：可配置容忍时差（默认 5ms），超出则拒绝生成
- ID 解析：可从 ID 中提取时间戳、工作节点 ID、序列号

详情请参见：[snake/README.md](./snake/README.md)

### 消息队列 (mq/rabbitmq)

基于 [RabbitMQ](https://www.rabbitmq.com/) 构建的高性能消息队列客户端，专为 go-zero 框架设计的分布式消息处理模块。

**核心能力**：
- Sender（生产者）：单连接持久心跳（30s），断线自动指数退避重连（最大 30s）
- Listener（消费者）：实现 go-zero `queue.MessageQueue` 接口，支持 QoS 预取控制、手动/自动 ACK
- 集成 OpenTelemetry 链路追踪：通过 HeaderCarrier 实现 Trace Context 跨服务传播
- 内置 Prometheus 指标：Sender 5 项 + Listener 9 项（发送/消费计数、耗时、字节数、重连、断连、在途消息等）
- 拦截器链：Recovery（panic 捕获）→ Trace → Prometheus → Logging
- 优雅停机：通过 go-zero proc 协调，等待在途消息处理完成

详情请参见：[mq/rabbitmq/README.md](./mq/rabbitmq/README.md)

### 分布式任务队列 (cron)

基于 [Asynq](https://github.com/hibiken/asynq) 构建的分布式任务队列系统，专为 go-zero 框架设计的定时任务和异步任务处理模块。

**核心能力**：
- 基于 [Asynq](https://github.com/hibiken/asynq) 的分布式任务队列
- Client（生产者）：支持立即/延时/定时推送，JSON 自动序列化，自动注入 Trace Context
- Server（消费者）：支持 Cron 表达式定时任务、并发控制（默认 CPU 核数）、优先级队列
- 支持多种 Redis 模式：单机、哨兵、集群，支持 TLS
- 中间件链：Recovery（panic 捕获不重试）→ Prometheus → Trace
- 内置 Prometheus 指标：消费计数、耗时、字节数、活跃 Worker、重试次数、panic 次数等
- 任务管理：支持任务取消（CancelTask）和重新调度（RescheduleTask）

详情请参见：[cron/README.md](./cron/README.md)

### 阿里云 API 网关客户端 (aliyun/gateway)

基于 go-zero httpc 封装的阿里云 API 网关 Go 客户端，自动完成 HMAC-SHA256 v1 签名。

**核心能力**：
- 自动完成阿里云 API 网关 v1 HMAC-SHA256 签名，注入全套 `X-Ca-*` 头（Key、Nonce、Timestamp、Signature-Method、Signature-Headers、Signature）
- `Do` 方法：结构化请求，自动映射 path/form/json/header 标签
- `DoRaw` 方法：原始字节请求，适用于文件上传、XML、纯文本等场景
- 自动计算 Content-MD5（form/multipart 除外）
- 底层集成 go-zero httpc，同一 Host 自动共享熔断器
- 支持通过 `WithClient` 注入自定义 `*http.Client`（超时、TLS、连接池）

详情请参见：[aliyun/gateway/README.md](./aliyun/gateway/README.md)

### Kong HMAC Auth 客户端 (kong/hmacauth)

基于 go-zero httpc 封装的 Kong HMAC Auth Go 客户端，自动完成 HMAC 签名，遵循 [Kong HMAC Auth 插件](https://developer.konghq.com/plugins/hmac-auth/) 官方规范。

**核心能力**：
- 支持 5 种 HMAC 算法：hmac-sha1 / sha224 / sha256 / sha384 / sha512
- `@request-target` 伪头签名，符合 Kong 3.x 官方规范
- 自动注入 Date（RFC 2822 UTC）、User-Agent、Host、Digest 头
- Digest 完整性：包含空 body 场景的 SHA-256 摘要计算（form/multipart 除外）
- `Do`/`DoRaw` 双方法：结构化请求 + 原始字节请求
- 可自定义参与签名的 header 列表，默认 `["date", "@request-target"]`
- 底层集成 go-zero httpc，同一 Host 自动共享熔断器
- 支持通过 `WithClient` 注入自定义 `*http.Client`

详情请参见：[kong/hmacauth/README.md](./kong/hmacauth/README.md)

## 贡献者

感谢所有为本项目做出贡献的开发者：

<a href="https://github.com/lerity-yao/czt-contrib/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=lerity-yao/czt-contrib" />
</a>

## 贡献与许可

欢迎通过 Issue 和 Pull Request 参与贡献。项目采用 MIT 许可证，详情请参见 [LICENSE](LICENSE)。
