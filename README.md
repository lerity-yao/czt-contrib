# czt-contrib 微服务治理组件库

## 项目概述

czt-contrib 是一个专注于微服务治理的 Go 语言组件库，提供了服务注册中心、配置中心以及分布式 ID 生成器等核心功能模块。该项目采用 Go 语言开发，兼容 go-zero 微服务框架，旨在为分布式系统提供一套完整的基础组件解决方案。

所有组件都能完美的嵌入到 go-zero 微服务框架中，提供一站式微服务治理解决方案。

## 项目结构

```
czt-contrib/
├── configcenter/      # 配置中心模块
│   ├── consul/       # Consul 配置中心实现
│   └── nacos/        # Nacos 配置中心实现（待实现）
├── registercenter/   # 服务注册中心模块
│   └── consul/       # Consul 服务注册实现
├── snake/            # 分布式ID生成器模块
├── go.mod           # Go 模块定义
├── go.sum           # Go 依赖校验
└── main.go          # 示例入口文件
```


## 功能模块

### 1. 配置中心 (configcenter)

#### Consul 配置中心
- 基于 HashiCorp Consul 实现的配置管理
- 支持动态配置更新
- 提供配置监听功能

详情请参见：[configcenter/consul/README.md](./configcenter/consul/README.md)

#### Nacos 配置中心
- **待实现**：基于 Alibaba Nacos 的配置管理（计划中）

### 2. 服务注册中心 (registercenter)

#### Consul 服务注册
- 服务注册与发现功能
- 健康检查机制
- 服务负载均衡支持

详情请参见：[registercenter/consul/README.md](./registercenter/consul/README.md)

### 3. 分布式 ID 生成器 (snake)

#### Snake 雪花算法
- 基于雪花算法的分布式唯一 ID 生成器
- 高性能、低延迟的 ID 生成
- 支持自动工作节点 ID 分配
- 内置时钟回拨处理机制

详情请参见：[snake/README.md](./snake/README.md)

## 技术特点

- **高可用性**：支持多种注册中心和配置中心实现
- **高性能**：优化的并发处理能力
- **易集成**：兼容主流微服务框架
- **可扩展**：模块化设计，易于扩展新功能


### 4. rabbitmq 消息队列

基于 [RabbitMQ](https://www.rabbitmq.com/) 构建的高性能消息队列客户端，专为 Go-Zero 框架设计的分布式消息处理模块。

## ✨ 特性

- 🚀 **高性能**: 基于 AMQP 协议的高性能消息队列
- 🔗 **链路追踪**: 集成 OpenTelemetry 分布式链路追踪
- 📊 **监控指标**: 内置 Prometheus 指标收集
- 🛡️ **错误恢复**: 自动重连和错误处理机制
- 🔄 **消息确认**: 支持可靠的消息确认机制
- ⚡ **并发控制**: 灵活的 QoS 配置和并发控制
- 🔧 **配置灵活**: 支持多种消息模式和队列配置

详情请参见：[mq/rabbitmq/README.md](./mq/rabbitmq/README.md)

### 5. cron 分布式任务队列系

基于 [Asynq](https://github.com/hibiken/asynq) 构建的分布式任务队列系统，专为 Go-Zero 框架设计的定时任务和异步任务处理模块。

## 特性

- 🚀 **高性能**: 基于 Redis 的高性能分布式任务队列
- ⏰ **定时任务**: 支持 Cron 表达式定时任务
- 🔄 **异步处理**: 异步任务队列，支持延迟执行
- 📊 **监控指标**: 内置 Prometheus 指标收集
- 🔍 **链路追踪**: 集成 OpenTelemetry 链路追踪
- 🛡️ **错误恢复**: 自动 panic 恢复和错误处理
- 🔧 **配置灵活**: 支持多种 Redis 模式（单机、哨兵、集群）

详情请参见：[cron/README.md](./cron/README.md)
