---
layout: home
hero:
  name: czt-contrib
  text: go-zero 微服务治理工具集
  tagline: 自 2023 年起生产验证 —— 注册中心、配置中心、消息队列、定时任务、分布式 ID、API 网关客户端、代码生成器。
  actions:
    - theme: brand
      text: 快速开始
      link: /zh/guide/getting-started
    - theme: alt
      text: 在 GitHub 上查看
      link: https://github.com/lerity-yao/czt-contrib
features:
  - icon: 🆔
    title: Snake
    details: 基于雪花算法的高性能分布式 ID 生成器，使用 FNV 哈希分配 Worker ID
    link: /zh/modules/snake
  - icon: ⏰
    title: Cron
    details: 基于 Asynq + Redis 的分布式定时任务调度，集成 OpenTelemetry 可观测性
    link: /zh/modules/cron
  - icon: 🔧
    title: 配置中心
    details: 基于 Consul 的配置管理，支持长轮询热更新
    link: /zh/modules/configcenter
  - icon: 📡
    title: 注册中心
    details: 基于 Consul 的服务注册，集成 gRPC Resolver
    link: /zh/modules/registercenter
  - icon: 📨
    title: RabbitMQ
    details: 消息队列客户端，支持拦截器链、链路追踪与指标上报
    link: /zh/modules/rabbitmq
  - icon: 🌐
    title: 阿里云网关
    details: 阿里云 API 网关签名客户端（X-Ca-* 请求头）
    link: /zh/modules/aliyun-gateway
  - icon: 🔐
    title: Kong HMAC 认证
    details: 符合 Kong 3.x 规范的 HMAC 认证客户端
    link: /zh/modules/kong-hmacauth
  - icon: 🛠️
    title: cztctl
    details: go-zero 项目代码生成器，支持 API、RPC SDK、Cron、MQ 脚手架
    link: /zh/modules/cztctl
---
