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
