# CZT-Contrib RabbitMQ Module

基于 [RabbitMQ](https://www.rabbitmq.com/) 构建的高性能消息队列客户端，专为 Go-Zero 框架设计的分布式消息处理模块。

## ✨ 特性

- 🚀 **高性能**: 基于 AMQP 协议的高性能消息队列
- 🔗 **链路追踪**: 集成 OpenTelemetry 分布式链路追踪
- 📊 **监控指标**: 内置 Prometheus 指标收集
- 🛡️ **错误恢复**: 自动重连和错误处理机制
- 🔄 **消息确认**: 支持可靠的消息确认机制
- ⚡ **并发控制**: 灵活的 QoS 配置和并发控制
- 🔧 **配置灵活**: 支持多种消息模式和队列配置

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

```go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lerity-yao/czt-contrib/mq/rabbitmq"
)

func main() {
	// 配置生产者
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

	// 创建生产者客户端
	ctx := context.Background()
	client, err := rabbitmq.MustNewSenderClient(ctx, conf)
	if err != nil {
		panic(err)
	}

	// 发送消息
	message := map[string]interface{}{
		"order_id": "12345",
		"amount":   99.99,
		"user_id":  "user001",
	}

	payload, _ := json.Marshal(message)
	
	// 发送到指定交换机和路由键
	err = client.Send(ctx, "order.exchange", "order.created", payload)
	if err != nil {
		panic(err)
	}

	fmt.Println("消息发送成功!")
}
```

### 2. 消费者（服务端）使用

```go
package main

import (
	"context"
	"encoding/json"
	"flag"

	"github.com/lerity-yao/czt-contrib/mq/rabbitmq"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/rest"
)

// 配置结构
type Config struct {
	rest.RestConf
	RabbitMqConf rabbitmq.RabbitListenerConf
}

// 消息结构
type OrderMessage struct {
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount"`
	UserID  string  `json:"user_id"`
}

// 消费者逻辑
type OrderProcessor struct {
	config Config
}

func NewOrderProcessor(c Config) *OrderProcessor {
	return &OrderProcessor{config: c}
}

func (p *OrderProcessor) Consume(ctx context.Context, message []byte) error {
	var order OrderMessage
	
	if err := json.Unmarshal(message, &order); err != nil {
		logc.Errorf(ctx, "消息解析失败: %v, 原始消息: %s", err, string(message))
		return err
	}

	logc.Infof(ctx, "处理订单: ID=%s, 金额=%.2f, 用户=%s", 
		order.OrderID, order.Amount, order.UserID)

	// 处理业务逻辑
	// 例如：更新数据库、发送通知等
	
	return nil
}

// 创建消费者服务
func CreateRabbitMQService(c Config) service.Service {
	logx.Infof("连接 RabbitMQ: %s:%d, 监听队列: %v",
		c.RabbitMqConf.Host, c.RabbitMqConf.Port,
		c.RabbitMqConf.ListenerQueues)

	ctx := context.Background()
	return rabbitmq.MustNewListener(ctx, c.RabbitMqConf, NewOrderProcessor(c))
}

func main() {
	flag.Parse()

	// 配置消费者队列
	queueConfig := rabbitmq.ConsumerConf{
		Name:          "order.queue",
		MaxRetryCount: 3,
		AutoAck:       false,  // 手动确认
		Exclusive:     false,  // 允许多个消费者
		NoLocal:       false,
		NoWait:        false,
	}

	queues := []rabbitmq.ConsumerConf{queueConfig}

	config := Config{
		RestConf: rest.RestConf{
			Host: "0.0.0.0",
			Port: 8080,
		},
		RabbitMqConf: rabbitmq.RabbitListenerConf{
			RabbitConf: rabbitmq.RabbitConf{
				Username: "guest",
				Password: "guest",
				Host:     "localhost",
				Port:     5672,
				VHost:    "/",
			},
			ListenerQueues: queues,
			ChannelQos: rabbitmq.ChannelQosConf{
				PrefetchCount: 10, // 每次预取10条消息
				PrefetchSize:  0,
				Global:        false,
			},
			ContentType: "application/json",
		},
	}

	// 设置服务
	if err := config.SetUp(); err != nil {
		panic(err)
	}

	// 启动服务组
	serviceGroup := service.NewServiceGroup()
	defer serviceGroup.Stop()

	// 添加RabbitMQ消费者服务
	serviceGroup.Add(CreateRabbitMQService(config))

	// 启动服务
	serviceGroup.Start()
}
```

## ⚙️ 配置说明

### RabbitMQ 基础配置

```go
type RabbitConf struct {
	Username string  // 用户名
	Password string  // 密码
	Host     string  // 主机地址
	Port     int     // 端口号
	VHost    string `json:",optional"` // 虚拟主机
}
```

### 生产者配置

```go
type RabbitSenderConf struct {
	RabbitConf
	ContentType string `json:",default=text/plain"` // 消息内容类型
}
```

### 消费者配置

```go
type RabbitListenerConf struct {
	RabbitConf
	ListenerQueues []ConsumerConf  // 监听队列配置
	ChannelQos     ChannelQosConf  // 通道QoS配置
	ContentType    string `json:",default=text/plain"` // 消息内容类型
}
```

### 队列消费配置

```go
type ConsumerConf struct {
	Name          string  // 队列名称
	MaxRetryCount int32 `json:",default=3"`    // 最大重试次数
	AutoAck       bool  `json:",default=false"` // 自动确认
	Exclusive     bool  `json:",default=false"` // 独占模式
	NoLocal       bool  `json:",default=false"` // 禁止本地消费
	NoWait        bool  `json:",default=false"` // 非阻塞模式
}
```

### 通道 QoS 配置

```go
type ChannelQosConf struct {
	PrefetchCount int  `json:",default=5"`   // 预取消息数量
	PrefetchSize  int  `json:",default=0"`   // 预取消息大小（字节）
	Global        bool `json:",default=false"` // 全局生效
}
```

## 🔧 高级功能

### 1. 链路追踪

模块自动集成 OpenTelemetry，支持分布式链路追踪：

- **生产者端**: 自动创建生产者 Span
- **消费者端**: 从消息头提取追踪信息
- **跨服务追踪**: 支持跨多个服务的完整调用链

### 2. 监控指标

内置 Prometheus 指标收集：

- `mq_consume_total`: 消息消费总数统计
- `mq_consume_duration_ms`: 消息消费耗时统计

### 3. 内置拦截器

模块提供以下内置拦截器：

- **RecoveryInterceptor**: 自动 panic 恢复
- **TraceInterceptor**: 链路追踪
- **PrometheusInterceptor**: 指标收集
- **LoggingInterceptor**: 日志记录

### 4. 自定义拦截器

```go
// 自定义拦截器示例
func customInterceptor(ctx context.Context, queueName string, message []byte, next func(context.Context, []byte) error) error {
	start := time.Now()
	
	// 前置处理
	logc.Infof(ctx, "开始处理队列 %s 的消息", queueName)
	
	// 执行下一个拦截器或业务逻辑
	err := next(ctx, message)
	
	// 后置处理
	duration := time.Since(start)
	if err != nil {
		logc.Errorf(ctx, "处理失败，耗时: %v, 错误: %v", duration, err)
	} else {
		logc.Infof(ctx, "处理成功，耗时: %v", duration)
	}
	
	return err
}

// 使用自定义拦截器
interceptor := rabbitmq.Chain(
	rabbitmq.RecoveryInterceptor,
	rabbitmq.TraceInterceptor,
	customInterceptor,  // 添加自定义拦截器
	rabbitmq.PrometheusInterceptor,
)
```

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