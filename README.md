# czt-contrib 微服务治理组件库

## 项目概述

czt-contrib 是一个专注于微服务治理的 Go 语言组件库，提供了服务注册中心、配置中心以及分布式 ID 生成器等核心功能模块。该项目采用 Go 语言开发，兼容 go-zero 微服务框架，旨在为分布式系统提供一套完整的基础组件解决方案。

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

## 快速开始

### 环境要求

- Go 1.23+
- Consul 服务（如使用 Consul 模块）

