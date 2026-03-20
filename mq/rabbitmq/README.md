# CZT-Contrib RabbitMQ Module

基于 [RabbitMQ](https://www.rabbitmq.com/) 构建的高性能消息队列客户端，专为 Go-Zero 框架设计的分布式消息处理模块。

## ✨ 特性

- 🚀 **高性能**: 基于 AMQP 协议的高性能消息队列
- 🔗 **链路追踪**: 集成 OpenTelemetry 分布式链路追踪
- 📊 **监控指标**: 内置 Prometheus 指标收集（Sender 5个 + Listener 9个）
- 🛡️ **错误恢复**: 自动重连和错误处理机制
- 🔄 **消息确认**: 支持可靠的消息确认机制
- ⚡ **并发控制**: 灵活的 QoS 配置和并发控制
- 🔧 **配置灵活**: 支持多种消息模式和队列配置
- 🔌 **连接可靠**: 支持自动重连、NotifyClose 监听、优雅停机

---

## 📦 安装

```bash
go get github.com/lerity-yao/czt-contrib/mq/rabbitmq
```

## 📦 服务生成

支持类似 goctl 工具一键生成服务端代码， 工具为 cztctl, 是goctl魔改的

也可以自定义服务端代码生成模板

请参考 https://github.com/lerity-yao/go-zero/tree/cztctl

## 🚀 快速开始

### 1. 生产者（客户端）使用

#### go-zero 框架中使用（推荐）

在 `svc/servicecontext.go` 中初始化 Sender，自动通过 `proc.AddShutdownListener` 优雅停机：

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
	"context"
	"your-project/internal/config"
	"github.com/lerity-yao/czt-contrib/mq/rabbitmq"
)

type ServiceContext struct {
	Config config.Config
	Sender rabbitmq.Sender
}

func NewServiceContext(c config.Config) *ServiceContext {
	ctx := context.Background()
	return &ServiceContext{
		Config: c,
		Sender: rabbitmq.MustNewSender(ctx, c.RabbitMqSenderConf),
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

#### 非 go-zero 框架使用

需要手动调用 `Close()` 关闭连接：

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

### 2. 消费者（服务端）使用

消费者仅支持在 go-zero 框架中使用，推荐使用 `cztctl` 工具生成代码。

**目录结构**：
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

**配置文件** `etc/demoa.yaml`：
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

**配置结构** `internal/config/config.go`：
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

**Handler 注册** `internal/handler/listeners.go`：
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

**Handler 实现** `internal/handler/demoA/gdemoahandler.go`：
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

**业务逻辑** `internal/logic/demoA/gdemoalogic.go`：
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

**主函数** `demoa.go`：
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

## ⚙️ 配置说明

### 1. Sender 配置 (RabbitSenderConf)

| 字段 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| Username | string | - | RabbitMQ 账号 |
| Password | string | - | RabbitMQ 密码 |
| Host | string | - | RabbitMQ 地址 |
| Port | int | - | RabbitMQ 端口 |
| VHost | string | "" | 虚拟主机（可选） |
| ContentType | string | text/plain | 消息内容类型 |

**YAML 配置示例**：
```yaml
RabbitMqSenderConf:
  Username: guest
  Password: guest
  Host: localhost
  Port: 5672
  VHost: /
  ContentType: application/json
```

### 2. Listener 配置 (RabbitListenerConf)

#### 基础配置

| 字段 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| Username | string | - | RabbitMQ 账号 |
| Password | string | - | RabbitMQ 密码 |
| Host | string | - | RabbitMQ 地址 |
| Port | int | - | RabbitMQ 端口 |
| VHost | string | "" | 虚拟主机（可选） |
| ContentType | string | text/plain | 消息内容类型 |
| ListenerQueues | []ConsumerConf | - | 监听队列配置列表 |
| ChannelQos | ChannelQosConf | - | 通道 QoS 配置 |

#### 队列消费配置 (ConsumerConf)

| 字段 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| Name | string | - | 队列名称 |
| AutoAck | bool | false | 自动确认。true=投递时自动确认（消息立即删除）；false=框架在消费完成后 ACK |
| Exclusive | bool | false | 独占模式。true=只允许当前消费者连接此队列 |
| NoLocal | bool | false | 禁止本地消费（RabbitMQ 不支持） |
| NoWait | bool | false | 非阻塞模式。true=不等待服务器响应 |

#### 通道 QoS 配置 (ChannelQosConf)

| 字段 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| PrefetchCount | int | 5 | 预取消息数量。未确认消息达到此上限时，RabbitMQ 停止投递新消息 |
| PrefetchSize | int | 0 | 预取消息总字节大小。0=不限制 |
| Global | bool | false | QoS 生效范围。false=仅当前消费者；true=影响所有消费者 |

**YAML 配置示例**：
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

## 🔧 高级功能

### 1. 链路追踪

模块自动集成 OpenTelemetry，支持分布式链路追踪：

- **生产者端**: 自动创建生产者 Span
- **消费者端**: 从消息头提取追踪信息
- **跨服务追踪**: 支持跨多个服务的完整调用链

### 2. 监控指标

内置 Prometheus 指标收集：

**Sender 指标**：
- `mq_sender_send_total`: 消息发送总数 (exchange, route_key, status)
- `mq_sender_send_duration_ms`: 消息发送耗时 (exchange, route_key)
- `mq_sender_send_size_bytes`: 消息发送大小 (exchange, route_key)
- `mq_sender_reconnect_total`: 重连次数
- `mq_sender_disconnect_total`: 掉线次数

**Listener 指标**：
- `mq_listener_consume_total`: 消息消费总数 (queue, status)
- `mq_listener_consume_duration_ms`: 消息消费耗时 (queue)
- `mq_listener_consume_size_bytes`: 消息消费大小 (queue)
- `mq_listener_in_flight`: 当前处理中消息数 (queue)
- `mq_listener_parse_error_total`: 解析失败数 (queue)
- `mq_listener_panic_total`: Panic 次数 (queue)
- `mq_listener_ack_total`: ACK/Reject 计数 (queue, type)
- `mq_listener_reconnect_total`: 重连次数
- `mq_listener_disconnect_total`: 掉线次数

### 3. 内置拦截器

模块提供以下内置拦截器：

- **RecoveryInterceptor**: 自动 panic 恢复
- **TraceInterceptor**: 链路追踪
- **PrometheusInterceptor**: 指标收集
- **LoggingInterceptor**: 日志记录


## 🏆 最佳实践

### 1. 消息设计

- **消息格式**: 使用 JSON 格式，便于序列化和调试
- **消息大小**: 控制消息大小，避免大消息影响性能
- **幂等性**: 确保消息处理是幂等的，支持重试

### 2. 队列配置

- **持久化**: 重要消息应设置持久化队列
- **死信队列**: 配置死信队列处理失败消息
- **TTL**: 设置合理的消息过期时间

### 3. 错误处理

- **重试机制**: 合理配置重试次数和间隔
- **死信处理**: 处理无法重试成功的消息
- **监控告警**: 设置消息积压和失败告警

### 4. 性能优化

- **QoS 配置**: 根据业务需求调整预取数量
- **并发控制**: 合理设置消费者并发数
- **连接复用**: 复用连接和通道，减少开销

## 🔍 故障排除

### 常见问题

1. **连接失败**
    - 检查 RabbitMQ 服务状态
    - 验证网络连接和防火墙设置
    - 确认用户名密码和权限

2. **消息丢失**
    - 确认消息持久化设置
    - 检查消费者确认机制
    - 验证死信队列配置

3. **性能问题**
    - 调整 QoS 预取设置
    - 优化消息大小和格式
    - 增加消费者实例


## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📞 联系方式

- 项目主页: [GitHub](https://github.com/lerity-yao/czt-contrib)
- 问题反馈: [Issues](https://github.com/lerity-yao/czt-contrib/issues)

## 📋 更新日志

查看 [CHANGELOG.md](./CHANGELOG.md)