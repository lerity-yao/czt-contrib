# rabbitmq

[English](./README.md)

基于 [amqp091-go](https://github.com/rabbitmq/amqp091-go) 构建的 RabbitMQ 客户端，专为 [go-zero](https://github.com/zeromicro/go-zero) 框架设计，内置**链路追踪**、**Prometheus 监控**、**自动重连**和**优雅停机**。

## 特性

- 📤 **Sender / Listener 双模式** — `Sender` 发送消息到 Exchange，`Listener` 消费 Queue 消息
- 🔗 **链路追踪** — 集成 OpenTelemetry，生产者自动注入 Span，消费者自动提取并续链
- 📊 **监控指标** — 内置 Prometheus 指标（Sender 5 个 + Listener 9 个），接入 go-zero `/metrics`
- 🔄 **自动重连** — 监听 Connection/Channel 的 NotifyClose，断线后自动重连（最多 10 次）
- 🛡️ **优雅停机** — Listener 先停消费再排空任务，Sender 通过 `proc.AddShutdownListener` 自动注册关闭钩子
- 🛡️ **Panic 恢复** — Listener 内置 Recovery 拦截器，panic 不会导致消费协程退出
- 🔧 **拦截器链** — Sender 和 Listener 各自支持拦截器链（`SenderChain` / `Chain`），可扩展自定义逻辑
- 🏗️ **Admin 管理端** — 声明 Exchange、Queue、绑定关系

## 安装

```bash
go get github.com/lerity-yao/czt-contrib/mq/rabbitmq
```

## 服务生成

支持类似 goctl 工具一键生成服务端代码，工具为 cztctl，是 goctl 魔改的。

也可以自定义服务端代码生成模板。

请参考 [cztctl](https://github.com/lerity-yao/czt-contrib/blob/main/cztctl/README.md)

## 配置参数

### RabbitConf（基础连接配置）

| 参数名 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `Username` | string | 是 | RabbitMQ 账号 |
| `Password` | string | 是 | RabbitMQ 密码 |
| `Host` | string | 是 | RabbitMQ 地址 |
| `Port` | int | 是 | RabbitMQ 端口 |
| `VHost` | string | 否 | 虚拟主机，默认为空 |

### RabbitSenderConf（Sender 配置）

继承 `RabbitConf`，额外字段：

| 参数名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `ContentType` | string | `text/plain` | 发送消息的 MIME 类型 |

### RabbitListenerConf（Listener 配置）

继承 `RabbitConf`，额外字段：

| 参数名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `ListenerQueues` | []ConsumerConf | — | 监听队列配置列表 |
| `ChannelQos` | ChannelQosConf | — | 通道 QoS 配置 |
| `ContentType` | string | `text/plain` | 重推消息的 MIME 类型 |

### ConsumerConf（队列消费配置）

| 参数名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `Name` | string | — | 队列名称 |
| `AutoAck` | bool | `false` | `true` = 投递时自动确认（消息立即删除，无重试机会）；`false` = 框架在消费完成后调用 `Ack(false)` 确认（无论成功或失败） |
| `Exclusive` | bool | `false` | 独占模式。`true` = 只允许当前消费者连接此队列 |
| `NoLocal` | bool | `false` | 禁止本地消费（RabbitMQ 不支持此模式） |
| `NoWait` | bool | `false` | 非阻塞模式。`true` = 不等待服务器响应 |

> `AutoAck=false` 时，框架层不处理重试，业务如需重试请在 handler 中自行实现（如发送延迟队列、记录 DB 等）。

### ChannelQosConf（通道 QoS 配置）

| 参数名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `PrefetchCount` | int | `5` | 预取消息数量。未确认消息达到此上限时，RabbitMQ 停止投递新消息 |
| `PrefetchSize` | int | `0` | 预取消息总字节大小。`0` = 不限制 |
| `Global` | bool | `false` | QoS 生效范围。`false` = 仅当前消费者；`true` = 影响所有消费者（建议 `false`） |

> 如果一个消费者监控多个队列，这些队列共享同一个 QoS 设置。

### ExchangeConf（Exchange 声明配置）

| 参数名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `ExchangeName` | string | — | Exchange 名称 |
| `Type` | string | — | Exchange 类型，可选：`direct`、`fanout`、`topic`、`headers` |
| `Durable` | bool | `true` | 是否持久化 |
| `AutoDelete` | bool | `false` | 是否自动删除 |
| `Internal` | bool | `false` | 是否为内部 Exchange |
| `NoWait` | bool | `false` | 是否不等待服务器响应 |
| `Queues` | []QueueConf | — | 绑定的队列配置列表 |

### QueueConf（Queue 声明配置）

| 参数名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `Name` | string | — | 队列名称 |
| `Durable` | bool | `true` | 是否持久化 |
| `AutoDelete` | bool | `false` | 是否自动删除 |
| `Exclusive` | bool | `false` | 是否独占 |
| `NoWait` | bool | `false` | 是否不等待服务器响应 |

## API 参考

### 构造函数

| 函数 | 签名 | 说明 |
|------|------|------|
| `MustNewSender` | `func MustNewSender(conf RabbitSenderConf) Sender` | 创建 Sender，失败 panic |
| `NewSender` | `func NewSender(conf RabbitSenderConf) (Sender, error)` | 创建 Sender，失败返回 error |
| `MustNewListener` | `func MustNewListener(conf RabbitListenerConf, handler ConsumeHandler) queue.MessageQueue` | 创建 Listener，失败 panic |
| `MustNewAdmin` | `func MustNewAdmin(conf RabbitConf) *Admin` | 创建 Admin，失败 panic |

> `NewSender` 内部通过 `proc.AddShutdownListener` 注册优雅关闭钩子，go-zero 环境下无需手动调用 `Close()`。
> `MustNewListener` 返回 `queue.MessageQueue` 接口，需调用 `Start()` 阻塞运行。

### Sender 接口

| 方法 | 签名 | 说明 |
|------|------|------|
| `Send` | `Send(ctx context.Context, exchange string, routeKey string, msg []byte) error` | 发送消息。连接断开时自动重连后再发送 |
| `Close` | `Close() error` | 关闭 Channel 和 Connection，标记 `closed` 防止重连 |

### Listener 方法

| 方法 | 签名 | 说明 |
|------|------|------|
| `Start` | `Start()` | 启动消费，阻塞运行（监听所有 `ListenerQueues`） |
| `Stop` | `Stop()` | 优雅停机：标记 `closed` → 关闭 Channel 停止消费 → 等待任务排空（最多 10s） → 关闭 Connection |

> `Listener` 实现了 `queue.MessageQueue` 接口（`Start` / `Stop`），可直接加入 go-zero `ServiceGroup`。

### Admin 方法

| 方法 | 签名 | 说明 |
|------|------|------|
| `DeclareExchange` | `DeclareExchange(conf ExchangeConf, args amqp.Table) error` | 声明 Exchange |
| `DeclareQueue` | `DeclareQueue(conf QueueConf, args amqp.Table) error` | 声明 Queue |
| `Bind` | `Bind(queueName string, routekey string, exchange string, notWait bool, args amqp.Table) error` | 绑定 Queue 到 Exchange |

### ConsumeHandler 接口

| 方法 | 签名 | 说明 |
|------|------|------|
| `Consume` | `Consume(ctx context.Context, message []byte) error` | 消费消息的业务逻辑 |

### HandlerFunc

| 类型 | 定义 | 说明 |
|------|------|------|
| `HandlerFunc` | `func(ctx context.Context, message []byte) error` | 函数式处理器，实现 `ConsumeHandler` 接口 |

```go
// 推荐用法：每次消费时创建新 Logic，携带正确的 ctx
handler := func(ctx context.Context, message []byte) error {
    l := demoA.NewGDemoALogic(ctx, svcCtx)
    return l.GDemoA(message)
}
listener := rabbitmq.MustNewListener(conf, rabbitmq.HandlerFunc(handler))
```

### 拦截器

#### Listener 拦截器

| 类型 | 定义 |
|------|------|
| `Interceptor` | `func(ctx context.Context, queueName string, message []byte, next func(context.Context, []byte) error) error` |

| 函数 | 说明 |
|------|------|
| `Chain(interceptors ...Interceptor) Interceptor` | 构造 Listener 拦截器链 |

**内置 Listener 拦截器**（默认执行顺序：Recovery → Prometheus → Logging → Trace）：

| 拦截器 | 说明 |
|--------|------|
| `recoveryInterceptor` | 捕获 panic 并转为 error，记录日志和 `panic_total` 指标 |
| `prometheusInterceptor` | 记录消费耗时、消息大小、消费结果（success/fail） |
| `loggingInterceptor` | 消费失败时记录错误日志 |
| `traceInterceptor` | 解析 `RabbitMsgBody`，提取 Carrier 续链，将业务消息传给 handler |

#### Sender 拦截器

| 类型 | 定义 |
|------|------|
| `SenderInterceptor` | `func(ctx context.Context, exchange, routeKey string, msg []byte, next SenderFunc) error` |
| `SenderFunc` | `func(ctx context.Context, msg []byte) error` |

| 函数 | 说明 |
|------|------|
| `SenderChain(interceptors ...SenderInterceptor) SenderInterceptor` | 构造 Sender 拦截器链 |

**内置 Sender 拦截器**（默认执行顺序：Prometheus → Logging → Trace）：

| 拦截器 | 说明 |
|--------|------|
| `senderPrometheusInterceptor` | 记录发送耗时、消息大小、发送结果（success/fail） |
| `senderLoggingInterceptor` | 发送失败记录错误日志，成功记录消息内容 |
| `senderTraceInterceptor` | 开启生产者 Span，将 trace 上下文注入 Carrier，包装为 `RabbitMsgBody` 后发送 |

### RabbitMsgBody（消息结构体）

| 字段 | 类型 | 说明 |
|------|------|------|
| `Carrier` | `*propagation.HeaderCarrier` | OpenTelemetry 链路追踪头部数据 |
| `Msg` | `[]byte` | 业务消息内容 |

> Sender 的 `senderTraceInterceptor` 自动将原始消息包装为 `RabbitMsgBody{Carrier, Msg}` 再发送；Listener 的 `traceInterceptor` 自动解析 `RabbitMsgBody`，提取 Carrier 续链，将 `Msg` 传给业务 handler。

### Trace 辅助函数

| 函数 | 签名 | 说明 |
|------|------|------|
| `StartProducerSpan` | `StartProducerSpan(ctx context.Context, exchange string, routeKey string) (context.Context, oteltrace.Span)` | 开启生产者 Span |
| `StartConsumerSpan` | `StartConsumerSpan(ctx context.Context, queueName string, carrier *propagation.HeaderCarrier) (context.Context, oteltrace.Span)` | 开启消费者 Span，从 Carrier 提取上游 trace 上下文 |
| `EndSpan` | `EndSpan(span oteltrace.Span, err error)` | 结束 Span，err 非 nil 时标记 Error 状态 |

## 进阶指南

### 链路追踪

模块自动集成 OpenTelemetry，实现生产者到消费者的完整链路追踪：

1. **生产者端**：`senderTraceInterceptor` 开启 `rabbitmq-producer` Span，通过 `otel.GetTextMapPropagator().Inject()` 将 trace 上下文注入 `HeaderCarrier`，与业务消息一起包装为 `RabbitMsgBody` 序列化后发送
2. **消费者端**：`traceInterceptor` 解析 `RabbitMsgBody`，通过 `otel.GetTextMapPropagator().Extract()` 从 Carrier 恢复上游 trace 上下文，开启 `rabbitmq-consumer` Span 续链
3. **Span 属性**：生产者 Span 包含 `messaging.system=rabbitmq`、`messaging.destination=exchange`、`messaging.operation=send`；消费者 Span 包含 `messaging.destination=queueName`、`messaging.operation=process`

> 链路追踪集成在 go-zero 框架中，需要在项目中开启 OpenTelemetry 功能。你可以在 Jaeger 或 Grafana Tempo 中看到从 API 请求到消息消费的完整时序图。

### 监控指标

监控指标已并入 go-zero 的 Prometheus 体系，通过 `/metrics` 端点暴露。

#### Sender 端指标

| 指标 | 类型 | 标签 | 说明 |
|------|------|------|------|
| `rabbitmq_sender_send_total` | Counter | exchange, route_key, status | 消息发送总数（status: success/fail） |
| `rabbitmq_sender_send_duration_ms` | Histogram | exchange, route_key | 消息发送耗时(ms) |
| `rabbitmq_sender_send_size_bytes` | Histogram | exchange, route_key | 消息发送大小(bytes) |
| `rabbitmq_sender_reconnect_total` | Counter | — | 重连次数 |
| `rabbitmq_sender_disconnect_total` | Counter | — | 掉线次数 |

#### Listener 端指标

| 指标 | 类型 | 标签 | 说明 |
|------|------|------|------|
| `rabbitmq_listener_consume_total` | Counter | queue, status | 消息消费总数（status: success/fail） |
| `rabbitmq_listener_consume_duration_ms` | Histogram | queue | 消息消费耗时(ms) |
| `rabbitmq_listener_consume_size_bytes` | Histogram | queue | 消息消费大小(bytes) |
| `rabbitmq_listener_in_flight` | Gauge | queue | 当前正在处理的消息数 |
| `rabbitmq_listener_parse_error_total` | Counter | queue | 消息解析失败次数 |
| `rabbitmq_listener_panic_total` | Counter | queue | 消费 Panic 次数 |
| `rabbitmq_listener_ack_total` | Counter | queue, type | ACK/Reject 计数（type: ack/reject） |
| `rabbitmq_listener_reconnect_total` | Counter | — | 重连次数 |
| `rabbitmq_listener_disconnect_total` | Counter | — | 掉线次数 |

> 这些指标需要在 go-zero 项目中开启 Prometheus 监控功能。

### 自动重连机制

Sender 和 Listener 都实现了相同的重连策略：

1. **监听断开事件**：通过 `conn.NotifyClose()` 和 `channel.NotifyClose()` 监听 Connection 和 Channel 的关闭事件
2. **自动重连**：收到关闭事件后自动调用 `reconnect()`，重新建立 Connection 和 Channel
3. **互斥保护**：通过 `sync.Mutex` 保证同一时间只有一个重连过程，避免重复重连
4. **最大重试**：每次 `connect()` 内部最多重试 10 次，每次间隔 2 秒
5. **停机保护**：`closed` 标志位（`atomic.Bool`）置为 `true` 后，不再触发重连

**Sender 特有**：`Send()` 方法在发送前检查连接状态，如果已断开会自动重连后再发送。

**Listener 特有**：重连成功后会自动重新启动消费协程（调用 `internalStart()`），恢复所有队列的消费。

### 优雅停机

#### Listener

`Stop()` 执行流程：

1. 标记 `closed = true`（阻止重连和接收新消息）
2. 关闭 Channel（让消费者 goroutine 自然退出）
3. `listenerWg.Wait()`（等待消费者协程退出——"先停水龙头"）
4. `taskWg.Wait()`（最多等 10 秒，等待正在处理的消息完成——"再排空水池"）
5. 关闭 Connection

#### Sender

`NewSender()` 内部自动注册 `proc.AddShutdownListener`，go-zero 优雅停机时自动调用 `Close()`：

1. 标记 `closed = true`（阻止重连）
2. 关闭 Channel
3. 关闭 Connection

> 非 go-zero 环境使用 `NewSender()` 创建时，需手动调用 `Close()` 关闭连接。

### 拦截器自定义

默认拦截器链已覆盖 Recovery / Prometheus / Logging / Trace 四个方面，一般无需自定义。如需扩展，可参考内置拦截器实现：

```go
// Listener 自定义拦截器
customInterceptor := func(ctx context.Context, queueName string, message []byte, next func(context.Context, []byte) error) error {
    // 前置逻辑
    err := next(ctx, message)
    // 后置逻辑
    return err
}

// Sender 自定义拦截器
customSenderInterceptor := func(ctx context.Context, exchange, routeKey string, msg []byte, next rabbitmq.SenderFunc) error {
    // 前置逻辑
    err := next(ctx, msg)
    // 后置逻辑
    return err
}
```

## 完整示例

### go-zero 集成：Sender（生产者）

```go
// internal/config/config.go
package config

import (
    "github.com/lerity-yao/czt-contrib/mq/rabbitmq"
    "github.com/zeromicro/go-zero/rest"
)

type Config struct {
    rest.RestConf
    RabbitMqSenderConf rabbitmq.RabbitSenderConf
}
```

```yaml
# etc/config.yaml
Name: order-api
Host: 0.0.0.0
Port: 8888

RabbitMqSenderConf:
  Username: guest
  Password: guest
  Host: localhost
  Port: 5672
  VHost: /
  ContentType: application/json
```

```go
// internal/svc/servicecontext.go
package svc

import (
    "your-project/internal/config"
    "github.com/lerity-yao/czt-contrib/mq/rabbitmq"
)

type ServiceContext struct {
    Config config.Config
    Sender rabbitmq.Sender
}

func NewServiceContext(c config.Config) *ServiceContext {
    return &ServiceContext{
        Config: c,
        Sender: rabbitmq.MustNewSender(c.RabbitMqSenderConf),
    }
}
```

```go
// internal/logic/createorderlogic.go
func (l *CreateOrderLogic) CreateOrder() error {
    payload, _ := json.Marshal(map[string]any{"order_id": "12345"})
    return l.svcCtx.Sender.Send(l.ctx, "order.exchange", "order.created", payload)
}
```

### go-zero 集成：Listener（消费者）

推荐使用 `cztctl` 工具生成代码。

**目录结构**

```
├── etc/
│   └── demoa.yaml
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── handler/
│   │   ├── demoA/
│   │   │   └── gdemoahandler.go
│   │   └── listeners.go
│   ├── logic/
│   │   └── demoA/
│   │       └── gdemoalogic.go
│   └── svc/
│       └── servicecontext.go
└── demoa.go
```

```yaml
# etc/demoa.yaml
Name: demoA
Host: 127.0.0.1
Port: 8080

GDemoARabbitmqConf:
  Username: guest
  Password: guest
  Host: localhost
  Port: 5672
  ListenerQueues:
    - Name: queue.demoa
      AutoAck: false
  ChannelQos:
    PrefetchCount: 5
    PrefetchSize: 0
    Global: false
```

```go
// internal/config/config.go
package config

import (
    "github.com/lerity-yao/czt-contrib/mq/rabbitmq"
    "github.com/zeromicro/go-zero/rest"
)

type Config struct {
    rest.RestConf
    GDemoARabbitmqConf rabbitmq.RabbitListenerConf
}
```

```go
// internal/handler/listeners.go
// Code generated by cztctl. DO NOT EDIT.
package handler

import (
    "your-project/internal/handler/demoA"
    "your-project/internal/svc"
    "github.com/zeromicro/go-zero/core/service"
)

func RegisterHandlers(server *service.ServiceGroup, serverCtx *svc.ServiceContext) {
    server.Add(demoA.GDemoAHandler(serverCtx))
}
```

```go
// internal/handler/demoA/gdemoahandler.go
package demoA

import (
    "context"
    "your-project/internal/logic/demoA"
    "your-project/internal/svc"
    "github.com/lerity-yao/czt-contrib/mq/rabbitmq"
    "github.com/zeromicro/go-zero/core/service"
)

func GDemoAHandler(svcCtx *svc.ServiceContext) service.Service {
    handler := func(ctx context.Context, message []byte) error {
        l := demoA.NewGDemoALogic(ctx, svcCtx)
        return l.GDemoA(message)
    }
    return rabbitmq.MustNewListener(svcCtx.Config.GDemoARabbitmqConf, rabbitmq.HandlerFunc(handler))
}
```

```go
// internal/logic/demoA/gdemoalogic.go
package demoA

import (
    "context"
    "your-project/internal/svc"
    "github.com/zeromicro/go-zero/core/logx"
)

type GDemoALogic struct {
    logx.Logger
    ctx    context.Context
    svcCtx *svc.ServiceContext
}

func NewGDemoALogic(ctx context.Context, svcCtx *svc.ServiceContext) *GDemoALogic {
    return &GDemoALogic{
        Logger: logx.WithContext(ctx),
        ctx:    ctx,
        svcCtx: svcCtx,
    }
}

func (l *GDemoALogic) GDemoA(message []byte) error {
    // todo: add your logic here
    l.Infof("收到消息: %s", string(message))
    return nil
}
```

```go
// demoa.go
package main

import (
    "flag"
    "fmt"
    "your-project/internal/config"
    "your-project/internal/handler"
    "your-project/internal/svc"
    "github.com/zeromicro/go-zero/core/conf"
    "github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/demoa.yaml", "the config file")

func main() {
    flag.Parse()

    var c config.Config
    conf.MustLoad(*configFile, &c, conf.UseEnv())
    if err := c.SetUp(); err != nil {
        panic(err)
    }

    ctx := svc.NewServiceContext(c)

    serviceGroup := service.NewServiceGroup()
    defer serviceGroup.Stop()

    handler.RegisterHandlers(serviceGroup, ctx)

    fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
    serviceGroup.Start()
}
```

### 独立 Sender 使用

非 go-zero 环境中需要手动调用 `Close()` 关闭连接：

```go
package main

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/lerity-yao/czt-contrib/mq/rabbitmq"
)

func main() {
    conf := rabbitmq.RabbitSenderConf{
        RabbitConf: rabbitmq.RabbitConf{
            Username: "guest",
            Password: "guest",
            Host:     "localhost",
            Port:     5672,
            VHost:    "/",
        },
        ContentType: "application/json",
    }

    sender, err := rabbitmq.NewSender(conf)
    if err != nil {
        panic(err)
    }
    defer sender.Close() // 非 go-zero 环境需手动关闭

    ctx := context.Background()
    payload, _ := json.Marshal(map[string]any{"order_id": "12345"})

    if err := sender.Send(ctx, "order.exchange", "order.created", payload); err != nil {
        panic(err)
    }

    fmt.Println("消息发送成功!")
}
```

### Admin 使用

声明 Exchange、Queue 和绑定关系：

```go
package main

import (
    "fmt"
    "github.com/lerity-yao/czt-contrib/mq/rabbitmq"
)

func main() {
    admin := rabbitmq.MustNewAdmin(rabbitmq.RabbitConf{
        Username: "guest",
        Password: "guest",
        Host:     "localhost",
        Port:     5672,
    })

    // 声明 Exchange
    err := admin.DeclareExchange(rabbitmq.ExchangeConf{
        ExchangeName: "order.exchange",
        Type:         "direct",
        Durable:      true,
    }, nil)
    if err != nil {
        panic(err)
    }

    // 声明 Queue
    err = admin.DeclareQueue(rabbitmq.QueueConf{
        Name:    "order.created",
        Durable: true,
    }, nil)
    if err != nil {
        panic(err)
    }

    // 绑定 Queue 到 Exchange
    err = admin.Bind("order.created", "order.created", "order.exchange", false, nil)
    if err != nil {
        panic(err)
    }

    fmt.Println("Exchange 和 Queue 声明成功!")
}
```

## 更新日志

查看 [CHANGELOG.md](./CHANGELOG.md)
