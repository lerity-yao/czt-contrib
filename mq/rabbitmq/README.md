# rabbitmq

English | [中文](./readme-cn.md)

A RabbitMQ client built on top of [amqp091-go](https://github.com/rabbitmq/amqp091-go), designed for the [go-zero](https://github.com/zeromicro/go-zero) framework, with built-in **distributed tracing**, **Prometheus monitoring**, **automatic reconnection**, and **graceful shutdown**.

## Features

- 📤 **Sender / Listener dual mode** — `Sender` publishes messages to an Exchange; `Listener` consumes messages from a Queue
- 🔗 **Distributed tracing** — Integrated with OpenTelemetry; producers automatically inject Spans, consumers automatically extract and continue the trace chain
- 📊 **Observability metrics** — Built-in Prometheus metrics (5 for Sender + 9 for Listener), exposed via go-zero `/metrics`
- 🔄 **Auto-reconnect** — Listens to `NotifyClose` on Connection/Channel and reconnects automatically (up to 10 retries)
- 🛡️ **Graceful shutdown** — Listener stops consuming first, then drains in-flight tasks; Sender registers a shutdown hook automatically via `proc.AddShutdownListener`
- 🛡️ **Panic recovery** — Listener has a built-in Recovery interceptor that converts panics to errors, preventing consumer goroutines from crashing
- 🔧 **Interceptor chain** — Both Sender and Listener support interceptor chains (`SenderChain` / `Chain`) for custom extensibility
- 🏗️ **Admin management** — Declare Exchanges, Queues, and bindings

## Installation

```bash
go get github.com/lerity-yao/czt-contrib/mq/rabbitmq
```

## Service Generation

Similar to the `goctl` tool, you can generate server-side code with a single command using `cztctl`, a customized fork of `goctl`.

Custom code generation templates are also supported.

See [cztctl](https://github.com/lerity-yao/czt-contrib/blob/main/cztctl/README.md) for details.

## Configuration

### RabbitConf (Base Connection Config)

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `Username` | string | Yes | RabbitMQ username |
| `Password` | string | Yes | RabbitMQ password |
| `Host` | string | Yes | RabbitMQ host address |
| `Port` | int | Yes | RabbitMQ port |
| `VHost` | string | No | Virtual host; defaults to empty |

### RabbitSenderConf (Sender Config)

Embeds `RabbitConf`, with the following additional field:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ContentType` | string | `text/plain` | MIME type of the published message |

### RabbitListenerConf (Listener Config)

Embeds `RabbitConf`, with the following additional fields:

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ListenerQueues` | []ConsumerConf | — | List of queue consumer configurations |
| `ChannelQos` | ChannelQosConf | — | Channel QoS configuration |
| `ContentType` | string | `text/plain` | MIME type used when requeuing messages |

### ConsumerConf (Queue Consumer Config)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Name` | string | — | Queue name |
| `AutoAck` | bool | `false` | `true` = auto-acknowledge on delivery (message is immediately deleted, no retry); `false` = the framework calls `Ack(false)` after consumption completes (regardless of success or failure) |
| `Exclusive` | bool | `false` | Exclusive mode. `true` = only the current consumer can connect to this queue |
| `NoLocal` | bool | `false` | Disable local consumption (not supported by RabbitMQ) |
| `NoWait` | bool | `false` | Non-blocking mode. `true` = do not wait for a server response |

> When `AutoAck=false`, the framework does not handle retries. If your business logic requires retries, implement them in the handler (e.g., send to a delay queue, persist to a database, etc.).

### ChannelQosConf (Channel QoS Config)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `PrefetchCount` | int | `5` | Number of messages to prefetch. RabbitMQ stops delivering new messages when the number of unacknowledged messages reaches this limit |
| `PrefetchSize` | int | `0` | Maximum total byte size of prefetched messages. `0` = no limit |
| `Global` | bool | `false` | QoS scope. `false` = current consumer only; `true` = applies to all consumers (recommended: `false`) |

> If a consumer listens to multiple queues, they all share the same QoS settings.

### ExchangeConf (Exchange Declaration Config)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `ExchangeName` | string | — | Exchange name |
| `Type` | string | — | Exchange type: `direct`, `fanout`, `topic`, or `headers` |
| `Durable` | bool | `true` | Whether the exchange is durable |
| `AutoDelete` | bool | `false` | Whether to auto-delete the exchange |
| `Internal` | bool | `false` | Whether this is an internal exchange |
| `NoWait` | bool | `false` | Whether to skip waiting for a server response |
| `Queues` | []QueueConf | — | List of queue configurations to bind |

### QueueConf (Queue Declaration Config)

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Name` | string | — | Queue name |
| `Durable` | bool | `true` | Whether the queue is durable |
| `AutoDelete` | bool | `false` | Whether to auto-delete the queue |
| `Exclusive` | bool | `false` | Whether the queue is exclusive |
| `NoWait` | bool | `false` | Whether to skip waiting for a server response |

## API Reference

### Constructors

| Function | Signature | Description |
|----------|-----------|-------------|
| `MustNewSender` | `func MustNewSender(conf RabbitSenderConf) Sender` | Creates a Sender; panics on failure |
| `NewSender` | `func NewSender(conf RabbitSenderConf) (Sender, error)` | Creates a Sender; returns an error on failure |
| `MustNewListener` | `func MustNewListener(conf RabbitListenerConf, handler ConsumeHandler) queue.MessageQueue` | Creates a Listener; panics on failure |
| `MustNewAdmin` | `func MustNewAdmin(conf RabbitConf) *Admin` | Creates an Admin; panics on failure |

> `NewSender` internally registers a graceful shutdown hook via `proc.AddShutdownListener`, so you do not need to call `Close()` manually in a go-zero environment.
> `MustNewListener` returns a `queue.MessageQueue` interface; call `Start()` to begin blocking execution.

### Sender Interface

| Method | Signature | Description |
|--------|-----------|-------------|
| `Send` | `Send(ctx context.Context, exchange string, routeKey string, msg []byte) error` | Sends a message. Automatically reconnects if the connection is down before sending |
| `Close` | `Close() error` | Closes the Channel and Connection, and sets the `closed` flag to prevent reconnection |

### Listener Methods

| Method | Signature | Description |
|--------|-----------|-------------|
| `Start` | `Start()` | Starts consuming messages; blocks until stopped (listens to all `ListenerQueues`) |
| `Stop` | `Stop()` | Graceful shutdown: sets `closed` → closes Channel to stop consuming → waits for tasks to drain (up to 10s) → closes Connection |

> `Listener` implements the `queue.MessageQueue` interface (`Start` / `Stop`) and can be added directly to a go-zero `ServiceGroup`.

### Admin Methods

| Method | Signature | Description |
|--------|-----------|-------------|
| `DeclareExchange` | `DeclareExchange(conf ExchangeConf, args amqp.Table) error` | Declares an Exchange |
| `DeclareQueue` | `DeclareQueue(conf QueueConf, args amqp.Table) error` | Declares a Queue |
| `Bind` | `Bind(queueName string, routekey string, exchange string, notWait bool, args amqp.Table) error` | Binds a Queue to an Exchange |

### ConsumeHandler Interface

| Method | Signature | Description |
|--------|-----------|-------------|
| `Consume` | `Consume(ctx context.Context, message []byte) error` | Business logic for consuming a message |

### HandlerFunc

| Type | Definition | Description |
|------|-----------|-------------|
| `HandlerFunc` | `func(ctx context.Context, message []byte) error` | A function-based handler that implements the `ConsumeHandler` interface |

```go
// Recommended usage: create a new Logic instance on each consumption with the correct ctx
handler := func(ctx context.Context, message []byte) error {
    l := demoA.NewGDemoALogic(ctx, svcCtx)
    return l.GDemoA(message)
}
listener := rabbitmq.MustNewListener(conf, rabbitmq.HandlerFunc(handler))
```

### Interceptors

#### Listener Interceptors

| Type | Definition |
|------|-----------|
| `Interceptor` | `func(ctx context.Context, queueName string, message []byte, next func(context.Context, []byte) error) error` |

| Function | Description |
|----------|-------------|
| `Chain(interceptors ...Interceptor) Interceptor` | Builds a Listener interceptor chain |

**Built-in Listener interceptors** (default execution order: Recovery → Prometheus → Logging → Trace):

| Interceptor | Description |
|-------------|-------------|
| `recoveryInterceptor` | Catches panics and converts them to errors; records a log entry and increments the `panic_total` metric |
| `prometheusInterceptor` | Records consumption latency, message size, and consumption result (success/fail) |
| `loggingInterceptor` | Logs an error when consumption fails |
| `traceInterceptor` | Parses `RabbitMsgBody`, extracts the Carrier to continue the trace chain, and passes the business message to the handler |

#### Sender Interceptors

| Type | Definition |
|------|-----------|
| `SenderInterceptor` | `func(ctx context.Context, exchange, routeKey string, msg []byte, next SenderFunc) error` |
| `SenderFunc` | `func(ctx context.Context, msg []byte) error` |

| Function | Description |
|----------|-------------|
| `SenderChain(interceptors ...SenderInterceptor) SenderInterceptor` | Builds a Sender interceptor chain |

**Built-in Sender interceptors** (default execution order: Prometheus → Logging → Trace):

| Interceptor | Description |
|-------------|-------------|
| `senderPrometheusInterceptor` | Records send latency, message size, and send result (success/fail) |
| `senderLoggingInterceptor` | Logs an error on send failure; logs the message content on success |
| `senderTraceInterceptor` | Starts a producer Span, injects the trace context into a Carrier, wraps the payload as `RabbitMsgBody`, and sends it |

### RabbitMsgBody (Message Struct)

| Field | Type | Description |
|-------|------|-------------|
| `Carrier` | `*propagation.HeaderCarrier` | OpenTelemetry trace propagation headers |
| `Msg` | `[]byte` | Business message payload |

> The Sender's `senderTraceInterceptor` automatically wraps the raw message as `RabbitMsgBody{Carrier, Msg}` before sending. The Listener's `traceInterceptor` automatically parses `RabbitMsgBody`, extracts the Carrier to continue the trace chain, and passes `Msg` to the business handler.

### Trace Helper Functions

| Function | Signature | Description |
|----------|-----------|-------------|
| `StartProducerSpan` | `StartProducerSpan(ctx context.Context, exchange string, routeKey string) (context.Context, oteltrace.Span)` | Starts a producer Span |
| `StartConsumerSpan` | `StartConsumerSpan(ctx context.Context, queueName string, carrier *propagation.HeaderCarrier) (context.Context, oteltrace.Span)` | Starts a consumer Span and extracts the upstream trace context from the Carrier |
| `EndSpan` | `EndSpan(span oteltrace.Span, err error)` | Ends a Span; marks it as an error if `err` is non-nil |

## Advanced Guide

### Distributed Tracing

The module integrates OpenTelemetry automatically to provide end-to-end tracing from producer to consumer:

1. **Producer side**: `senderTraceInterceptor` starts a `rabbitmq-producer` Span, injects the trace context into a `HeaderCarrier` via `otel.GetTextMapPropagator().Inject()`, then wraps it together with the business message into `RabbitMsgBody` and sends it
2. **Consumer side**: `traceInterceptor` parses `RabbitMsgBody`, restores the upstream trace context from the Carrier via `otel.GetTextMapPropagator().Extract()`, and starts a `rabbitmq-consumer` Span to continue the chain
3. **Span attributes**: The producer Span includes `messaging.system=rabbitmq`, `messaging.destination=exchange`, and `messaging.operation=send`; the consumer Span includes `messaging.destination=queueName` and `messaging.operation=process`

> Tracing integrates with the go-zero framework and requires OpenTelemetry to be enabled in your project. You can view the complete trace timeline from an API request through message consumption in Jaeger or Grafana Tempo.

### Observability Metrics

All metrics are integrated into go-zero's Prometheus system and exposed via the `/metrics` endpoint.

#### Sender Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `rabbitmq_sender_send_total` | Counter | exchange, route_key, status | Total messages sent (status: success/fail) |
| `rabbitmq_sender_send_duration_ms` | Histogram | exchange, route_key | Message send latency (ms) |
| `rabbitmq_sender_send_size_bytes` | Histogram | exchange, route_key | Message send size (bytes) |
| `rabbitmq_sender_reconnect_total` | Counter | — | Number of reconnections |
| `rabbitmq_sender_disconnect_total` | Counter | — | Number of disconnections |

#### Listener Metrics

| Metric | Type | Labels | Description |
|--------|------|--------|-------------|
| `rabbitmq_listener_consume_total` | Counter | queue, status | Total messages consumed (status: success/fail) |
| `rabbitmq_listener_consume_duration_ms` | Histogram | queue | Message consumption latency (ms) |
| `rabbitmq_listener_consume_size_bytes` | Histogram | queue | Message consumption size (bytes) |
| `rabbitmq_listener_in_flight` | Gauge | queue | Number of messages currently being processed |
| `rabbitmq_listener_parse_error_total` | Counter | queue | Number of message parse failures |
| `rabbitmq_listener_panic_total` | Counter | queue | Number of consumer panics |
| `rabbitmq_listener_ack_total` | Counter | queue, type | ACK/Reject count (type: ack/reject) |
| `rabbitmq_listener_reconnect_total` | Counter | — | Number of reconnections |
| `rabbitmq_listener_disconnect_total` | Counter | — | Number of disconnections |

> These metrics require Prometheus monitoring to be enabled in your go-zero project.

### Auto-Reconnect Mechanism

Both Sender and Listener implement the same reconnection strategy:

1. **Listen for disconnect events**: Monitors Connection and Channel close events via `conn.NotifyClose()` and `channel.NotifyClose()`
2. **Auto-reconnect**: Calls `reconnect()` automatically upon receiving a close event to re-establish the Connection and Channel
3. **Mutex protection**: Uses `sync.Mutex` to ensure only one reconnection process runs at a time, preventing duplicate reconnections
4. **Maximum retries**: Each `connect()` call retries up to 10 times internally, with a 2-second interval between attempts
5. **Shutdown guard**: Once the `closed` flag (`atomic.Bool`) is set to `true`, no further reconnection is triggered

**Sender-specific**: The `Send()` method checks the connection state before sending and reconnects automatically if the connection is down.

**Listener-specific**: After a successful reconnection, consumer goroutines are automatically restarted (by calling `internalStart()`) to resume consumption on all queues.

### Graceful Shutdown

#### Listener

`Stop()` execution flow:

1. Set `closed = true` (prevents reconnection and stops accepting new messages)
2. Close the Channel (allows consumer goroutines to exit naturally)
3. `listenerWg.Wait()` (waits for consumer goroutines to exit — "turn off the tap first")
4. `taskWg.Wait()` (waits up to 10 seconds for in-flight messages to complete — "then drain the pool")
5. Close the Connection

#### Sender

`NewSender()` automatically registers a `proc.AddShutdownListener` hook, so `Close()` is called automatically during go-zero's graceful shutdown:

1. Set `closed = true` (prevents reconnection)
2. Close the Channel
3. Close the Connection

> If you create a Sender with `NewSender()` outside a go-zero environment, you must call `Close()` manually to release the connection.

### Custom Interceptors

The default interceptor chain already covers Recovery, Prometheus, Logging, and Trace — no customization is needed in most cases. If you need to extend it, refer to the built-in interceptor implementations:

```go
// Custom Listener interceptor
customInterceptor := func(ctx context.Context, queueName string, message []byte, next func(context.Context, []byte) error) error {
    // pre-processing logic
    err := next(ctx, message)
    // post-processing logic
    return err
}

// Custom Sender interceptor
customSenderInterceptor := func(ctx context.Context, exchange, routeKey string, msg []byte, next rabbitmq.SenderFunc) error {
    // pre-processing logic
    err := next(ctx, msg)
    // post-processing logic
    return err
}
```

## Full Examples

### go-zero Integration: Sender (Producer)

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

### go-zero Integration: Listener (Consumer)

It is recommended to use the `cztctl` tool to generate the code.

**Directory structure**

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
    l.Infof("received message: %s", string(message))
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

### Standalone Sender Usage

Outside of a go-zero environment, you must call `Close()` manually to release the connection:

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
    defer sender.Close() // must be called manually outside go-zero

    ctx := context.Background()
    payload, _ := json.Marshal(map[string]any{"order_id": "12345"})

    if err := sender.Send(ctx, "order.exchange", "order.created", payload); err != nil {
        panic(err)
    }

    fmt.Println("Message sent successfully!")
}
```

### Admin Usage

Declare Exchanges, Queues, and bindings:

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

    // Declare Exchange
    err := admin.DeclareExchange(rabbitmq.ExchangeConf{
        ExchangeName: "order.exchange",
        Type:         "direct",
        Durable:      true,
    }, nil)
    if err != nil {
        panic(err)
    }

    // Declare Queue
    err = admin.DeclareQueue(rabbitmq.QueueConf{
        Name:    "order.created",
        Durable: true,
    }, nil)
    if err != nil {
        panic(err)
    }

    // Bind Queue to Exchange
    err = admin.Bind("order.created", "order.created", "order.exchange", false, nil)
    if err != nil {
        panic(err)
    }

    fmt.Println("Exchange and Queue declared successfully!")
}
```

## Changelog

See [CHANGELOG.md](./CHANGELOG.md)
# CZT-Contrib RabbitMQ Module

[中文](./readme-cn.md)

A high-performance message queue client built on [RabbitMQ](https://www.rabbitmq.com/), designed as a distributed message processing module for the Go-Zero framework.

## ✨ Features

- 🚀 **High Performance**: High-performance message queue based on the AMQP protocol
- 🔗 **Distributed Tracing**: Integrated OpenTelemetry distributed tracing
- 📊 **Metrics**: Built-in Prometheus metrics collection (5 Sender + 9 Listener metrics)
- 🛡️ **Error Recovery**: Automatic reconnection and error handling mechanism
- 🔄 **Message Acknowledgment**: Supports reliable message acknowledgment
- ⚡ **Concurrency Control**: Flexible QoS configuration and concurrency control
- 🔧 **Flexible Configuration**: Supports multiple message patterns and queue configurations
- 🔌 **Reliable Connection**: Supports automatic reconnection, NotifyClose listener, and graceful shutdown

---

## 📦 Installation

```bash
go get github.com/lerity-yao/czt-contrib/mq/rabbitmq
```

## 📦 Service Generation

Supports one-click server code generation similar to the goctl tool. The tool is `cztctl`, a customized version of goctl.

You can also customize the server code generation templates.

Please refer to [cztctl](https://github.com/lerity-yao/czt-contrib/blob/main/cztctl/README.md).

## 🚀 Quick Start

### 1. Producer (Client) Usage

#### Using in the go-zero Framework (Recommended)

Initialize the Sender in `svc/servicecontext.go`; graceful shutdown is handled automatically via `proc.AddShutdownListener`:

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
// internal/logic/xxxlogic.go
func (l *XxxLogic) Xxx() error {
	payload, _ := json.Marshal(message)
	return l.svcCtx.Sender.Send(l.ctx, "exchange", "routeKey", payload)
}
```

#### Using Outside the go-zero Framework

You need to call `Close()` manually to close the connection:

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
	payload, _ := json.Marshal(map[string]interface{}{"order_id": "12345"})
	
	if err := sender.Send(ctx, "order.exchange", "order.created", payload); err != nil {
		panic(err)
	}

	fmt.Println("消息发送成功!")
}
```

### 2. Consumer (Server) Usage

Consumers are only supported in the go-zero framework. It is recommended to generate code with the `cztctl` tool.

**Directory Structure**:
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
│   ├── logic/demoA/
│   │   └── gdemoalogic.go
│   └── svc/
│       └── servicecontext.go
└── demoa.go
```

**Configuration File** `etc/demoa.yaml`:
```yaml
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
```

**Configuration Struct** `internal/config/config.go`:
```go
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

**Handler Registration** `internal/handler/listeners.go`:
```go
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

**Handler Implementation** `internal/handler/demoA/gdemoahandler.go`:
```go
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

**Business Logic** `internal/logic/demoA/gdemoalogic.go`:
```go
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

**Main Function** `demoa.go`:
```go
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

## ⚙️ Configuration

### 1. Sender Configuration (RabbitSenderConf)

| Field | Type | Default | Description |
|--------|------|--------|------|
| Username | string | - | RabbitMQ username |
| Password | string | - | RabbitMQ password |
| Host | string | - | RabbitMQ host |
| Port | int | - | RabbitMQ port |
| VHost | string | "" | Virtual host (optional) |
| ContentType | string | text/plain | Message content type |

**YAML Configuration Example**:
```yaml
RabbitMqSenderConf:
  Username: guest
  Password: guest
  Host: localhost
  Port: 5672
  VHost: /
  ContentType: application/json
```

### 2. Listener Configuration (RabbitListenerConf)

#### Basic Configuration

| Field | Type | Default | Description |
|--------|------|--------|------|
| Username | string | - | RabbitMQ username |
| Password | string | - | RabbitMQ password |
| Host | string | - | RabbitMQ host |
| Port | int | - | RabbitMQ port |
| VHost | string | "" | Virtual host (optional) |
| ContentType | string | text/plain | Message content type |
| ListenerQueues | []ConsumerConf | - | List of listening queue configurations |
| ChannelQos | ChannelQosConf | - | Channel QoS configuration |

#### Queue Consumer Configuration (ConsumerConf)

| Field | Type | Default | Description |
|--------|------|--------|------|
| Name | string | - | Queue name |
| AutoAck | bool | false | Auto acknowledgment. true = auto-ack on delivery (message deleted immediately); false = framework ACK after consumption completes |
| Exclusive | bool | false | Exclusive mode. true = only the current consumer can connect to this queue |
| NoLocal | bool | false | Prohibit local consumption (not supported by RabbitMQ) |
| NoWait | bool | false | Non-blocking mode. true = do not wait for server response |

#### Channel QoS Configuration (ChannelQosConf)

| Field | Type | Default | Description |
|--------|------|--------|------|
| PrefetchCount | int | 5 | Number of prefetched messages. RabbitMQ stops delivering new messages when unacknowledged messages reach this limit |
| PrefetchSize | int | 0 | Total prefetched message size in bytes. 0 = unlimited |
| Global | bool | false | QoS scope. false = current consumer only; true = affects all consumers |

**YAML Configuration Example**:
```yaml
GDemoARabbitmqConf:
  Username: guest
  Password: guest
  Host: localhost
  Port: 5672
  VHost: /
  ContentType: application/json
  ListenerQueues:
    - Name: queue.demoa
      AutoAck: false
      Exclusive: false
    - Name: queue.demob
  ChannelQos:
    PrefetchCount: 10
    PrefetchSize: 0
    Global: false
```

## 🔧 Advanced Features

### 1. Distributed Tracing

The module integrates OpenTelemetry out of the box and supports distributed tracing:

- **Producer Side**: Automatically creates producer spans
- **Consumer Side**: Extracts trace information from message headers
- **Cross-Service Tracing**: Supports complete call chains across multiple services

### 2. Metrics

Built-in Prometheus metrics collection:

**Sender Metrics**:
- `rabbitmq_sender_send_total`: Total messages sent (exchange, route_key, status)
- `rabbitmq_sender_send_duration_ms`: Message send duration (exchange, route_key)
- `rabbitmq_sender_send_size_bytes`: Message send size (exchange, route_key)
- `rabbitmq_sender_reconnect_total`: Number of reconnections
- `rabbitmq_sender_disconnect_total`: Number of disconnections

**Listener Metrics**:
- `rabbitmq_listener_consume_total`: Total messages consumed (queue, status)
- `rabbitmq_listener_consume_duration_ms`: Message consumption duration (queue)
- `rabbitmq_listener_consume_size_bytes`: Message consumption size (queue)
- `rabbitmq_listener_in_flight`: Number of messages currently being processed (queue)
- `rabbitmq_listener_parse_error_total`: Number of parse failures (queue)
- `rabbitmq_listener_panic_total`: Number of panics (queue)
- `rabbitmq_listener_ack_total`: ACK/Reject count (queue, type)
- `rabbitmq_listener_reconnect_total`: Number of reconnections
- `rabbitmq_listener_disconnect_total`: Number of disconnections

### 3. Built-in Interceptors

The module provides the following built-in interceptors:

- **RecoveryInterceptor**: Automatic panic recovery
- **TraceInterceptor**: Distributed tracing
- **PrometheusInterceptor**: Metrics collection
- **LoggingInterceptor**: Logging


## 🏆 Best Practices

### 1. Message Design

- **Message Format**: Use JSON for easy serialization and debugging
- **Message Size**: Keep message sizes under control to avoid performance impact
- **Idempotency**: Ensure message processing is idempotent and supports retries

### 2. Queue Configuration

- **Durability**: Important messages should use durable queues
- **Dead Letter Queue**: Configure dead-letter queues to handle failed messages
- **TTL**: Set reasonable message expiration times

### 3. Error Handling

- **Retry Mechanism**: Configure retry counts and intervals reasonably
- **Dead Letter Handling**: Handle messages that cannot be retried successfully
- **Monitoring and Alerting**: Set up alerts for message backlog and failures

### 4. Performance Optimization

- **QoS Configuration**: Adjust prefetch counts based on business requirements
- **Concurrency Control**: Set consumer concurrency appropriately
- **Connection Reuse**: Reuse connections and channels to reduce overhead

## 🔍 Troubleshooting

### Common Issues

1. **Connection Failure**
    - Check RabbitMQ service status
    - Verify network connectivity and firewall settings
    - Confirm username, password, and permissions

2. **Message Loss**
    - Confirm message durability settings
    - Check consumer acknowledgment mechanism
    - Verify dead-letter queue configuration

3. **Performance Issues**
    - Adjust QoS prefetch settings
    - Optimize message size and format
    - Increase consumer instances


## 🤝 Contributing

Issues and Pull Requests are welcome!

## 📞 Contact

- Project Home: [GitHub](https://github.com/lerity-yao/czt-contrib)
- Issue Feedback: [Issues](https://github.com/lerity-yao/czt-contrib/issues)

## 📋 Changelog

See [CHANGELOG.md](./CHANGELOG.md)
