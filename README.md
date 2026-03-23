# czt-contrib 微服务治理组件库

## 项目概述

czt-contrib 是一个专注于微服务治理的 Go 语言组件库，提供了服务注册中心、配置中心、分布式 ID 生成器、消息队列、定时任务等核心功能模块。该项目采用 Go 语言开发，兼容 go-zero 微服务框架，旨在为分布式系统提供一套完整的基础组件解决方案。

所有组件都能完美的嵌入到 go-zero 微服务框架中，提供一站式微服务治理解决方案。

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
├── go.mod            # Go 模块定义
├── go.sum            # Go 依赖校验
└── main.go           # 示例入口文件
```

## 功能模块

### 代码生成工具 (cztctl)

基于 goctl 魔改的代码生成工具，专为 czt-contrib 组件库设计，支持一键生成 RabbitMQ 消费者服务、Cron 定时任务服务、Swagger 文档等。

**核心能力**：
- `cztctl api rabbitmq` - 根据 .api 文件生成 RabbitMQ 消费者服务代码
- `cztctl api cron` - 根据 .api 文件生成 Cron 定时任务服务代码
- `cztctl api swagger` - 根据 .api 文件生成 Swagger API 文档
- 支持自定义模板，可根据团队规范定制生成代码风格
- 生成代码自动集成链路追踪、监控指标、错误处理等能力

**IDE 插件**：
- [cztctl-intellij](https://github.com/lerity-yao/cztctl-intellij) - GoLand/IntelliJ IDEA 语法高亮插件，支持 `.cron`、`.rabbitmq` 文件的语法高亮、代码补全、错误检查

详情请参见：[cztctl/README.md](./cztctl/README.md)

### 配置中心 (configcenter)

基于 HashiCorp Consul 实现的分布式配置管理中心，专为 go-zero 框架设计。

**核心能力**：
- 基于 Consul KV 存储的配置管理
- 支持配置热更新，无需重启服务
- 提供配置变更监听回调，实时响应配置变化
- 支持多环境配置隔离（dev/test/prod）
- 与 go-zero 配置体系无缝集成

详情请参见：[configcenter/consul/README.md](./configcenter/consul/README.md)

### 服务注册中心 (registercenter)

基于 HashiCorp Consul 实现的服务注册与发现中心，专为 go-zero 框架设计。

**核心能力**：
- 服务自动注册与发现
- 多种健康检查机制（TTL、HTTP、gRPC）
- 服务负载均衡支持
- 服务下线自动摘除
- 支持 gRPC 和 HTTP 服务注册

详情请参见：[registercenter/consul/README.md](./registercenter/consul/README.md)

### 分布式 ID 生成器 (snake)

基于雪花算法实现的分布式唯一 ID 生成器，适用于高并发场景。

**核心能力**：
- 基于雪花算法，生成全局唯一、有序的 64 位 ID
- 高性能：单节点每秒可生成数百万个 ID
- 支持自动工作节点 ID 分配（基于 Consul）
- 内置时钟回拨检测与处理机制
- ID 可解析，含时间戳、机器标识、序列号

详情请参见：[snake/README.md](./snake/README.md)

### 消息队列 (mq/rabbitmq)

基于 [RabbitMQ](https://www.rabbitmq.com/) 构建的高性能消息队列客户端，专为 go-zero 框架设计的分布式消息处理模块。

**核心能力**：
- 基于 AMQP 0-9-1 协议，支持多种消息模式（Direct、Fanout、Topic）
- Sender（生产者）：支持同步/异步发送、消息确认、自动重连
- Listener（消费者）：支持 QoS 控制、手动/自动 ACK、并发消费
- 集成 OpenTelemetry 分布式链路追踪，生产-消费全链路可观测
- 内置 Prometheus 指标：发送/消费计数、耗时、字节数等
- 支持 cztctl 一键生成消费者服务代码

详情请参见：[mq/rabbitmq/README.md](./mq/rabbitmq/README.md)

### 分布式任务队列 (cron)

基于 [Asynq](https://github.com/hibiken/asynq) 构建的分布式任务队列系统，专为 go-zero 框架设计的定时任务和异步任务处理模块。

**核心能力**：
- 基于 Redis 的高性能分布式任务队列
- Server（消费者）：支持 Cron 表达式定时任务、并发控制、优先级队列
- Client（生产者）：支持立即执行、延时执行、定时执行、任务撤回
- 集成 OpenTelemetry 链路追踪，生产-消费全链路可观测
- 内置 Prometheus 指标：消费计数、耗时、队列状态、并发数等
- 自动 panic 恢复，panic 不重试，直接归档
- 支持多种 Redis 模式（单机、哨兵、集群）
- 支持 cztctl 一键生成定时任务服务代码

详情请参见：[cron/README.md](./cron/README.md)
