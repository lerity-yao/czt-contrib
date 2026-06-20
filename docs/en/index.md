---
layout: home
hero:
  name: czt-contrib
  text: Governance Toolkit for go-zero
  tagline: Production-proven since 2023 — registry, config center, MQ, cron, distributed ID, API gateway clients, and code generator.
  actions:
    - theme: brand
      text: Get Started
      link: /en/guide/getting-started
    - theme: alt
      text: View on GitHub
      link: https://github.com/lerity-yao/czt-contrib
features:
  - icon: 🆔
    title: Snake
    details: High-performance distributed ID generator based on Snowflake with FNV hash worker ID allocation
    link: /en/modules/snake
  - icon: ⏰
    title: Cron
    details: Distributed task scheduling based on Asynq + Redis with OpenTelemetry observability
    link: /en/modules/cron
  - icon: 🔧
    title: ConfigCenter
    details: Consul-based configuration management with long-polling hot reload
    link: /en/modules/configcenter
  - icon: 📡
    title: RegisterCenter
    details: Consul service registry with gRPC Resolver integration
    link: /en/modules/registercenter
  - icon: 📨
    title: RabbitMQ
    details: Message queue client with interceptor chain, tracing, and metrics
    link: /en/modules/rabbitmq
  - icon: 🌐
    title: Aliyun Gateway
    details: Alibaba Cloud API Gateway signature client (X-Ca-* headers)
    link: /en/modules/aliyun-gateway
  - icon: 🔐
    title: Kong HMAC Auth
    details: Kong HMAC authentication client compliant with Kong 3.x spec
    link: /en/modules/kong-hmacauth
  - icon: 🛠️
    title: cztctl
    details: Code generator for go-zero projects — API, RPC SDK, cron, MQ scaffolding
    link: /en/modules/cztctl
---
