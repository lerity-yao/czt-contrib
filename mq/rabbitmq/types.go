package rabbitmq

import (
	"context"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel/propagation"
)

// RabbitMsgBody 消息结构体
// RabbitMsgBody Carrier OpenTelemetry 链路跟踪 头部数据,注入连裤跟踪数据据
// RabbitMsgBody Msg 内容
type RabbitMsgBody struct {
	Carrier *propagation.HeaderCarrier
	Msg     []byte
}

// RabbitConf rabbitmq基础配置信息
// RabbitConf Username 账号
// RabbitConf Password 密码
// RabbitConf Host 地址
// RabbitConf Port 端口
// RabbitConf VHost 命名空间
type RabbitConf struct {
	Username string
	Password string
	Host     string
	Port     int
	VHost    string `json:",optional"`
}

// RabbitListener 消费者服务端结构体
// RabbitListener conn AMQP连接
// RabbitListener channel 通道
// RabbitListener forever 通到阻塞标志
// RabbitListener handler 允许客户端注入的消费逻辑
// RabbitListener queues 队列
// RabbitListener ctx 上下文
// RabbitListener maxRetry 服务端端口之后会重连，每次重连的最大次数
// RabbitListener reconnectMutex 重连锁
type RabbitListener struct {
	conn           *amqp.Connection
	channel        *amqp.Channel
	forever        chan bool
	handler        ConsumeHandler
	queues         RabbitListenerConf
	rootCtx        context.Context
	currentCtx     context.Context
	cancel         context.CancelFunc // 新增：用于通知协程停止
	taskWg         sync.WaitGroup     // 跟踪 processMessage
	listenerWg     sync.WaitGroup     // 跟踪 internalStart 里的协程
	maxRetry       int
	reconnectMutex sync.Mutex
	interceptor    Interceptor
}

// ChannelQosConf 通道qos设置
// ChannelQosConf 此处为消费者通到，非队列
// ChannelQosConf 如果一个消费者监控的多个队列，再此消费者中，所有队列的qos都是一样的
// ChannelQosConf PrefetchCount 可以预取的消息数量，当消费者未确认消息达到此数量上限，RabbitMQ 会停止向该消费者投递新消息，默认只为5，一次性最多处理5条消息
// ChannelQosConf PrefetchSize 可以预取的消息总数的字节总大小。设置为0，则不限制消息字节大小
// ChannelQosConf Global 此qos的生效范围，设置false，qos只在此消费者生效，如果设置true，会影响其他消费者，建议设置false
type ChannelQosConf struct {
	PrefetchCount int  `json:",default=5"`
	PrefetchSize  int  `json:",default=0"`
	Global        bool `json:",default=false"`
}

// ConsumerConf 队列消费配置参数
// ConsumerConf Name  消息队列名称
// ConsumerConf MaxRetryCount 消费消息重试的最大次数
// ConsumerConf AutoAck 控制消息确认机制,设置为true,则在消息被消费者拿到之后，队列会立马删除消息，通常要设置为false，在消费成功之后调用Ack(false)来通知队列删除消息，需要配合channel.Qos使用
// ConsumerConf Exclusive 队列访问控制模式，当前消费者是否唯一模式。设置true,则只允许当前消费者连接此队列，不允许其他消费者连接，设置false，则允许多个消费者连接队列。
// ConsumerConf NoLocal 禁止本地消费，即禁止消费者消费自己推送的消息；rabbitmq不支持此模式
// ConsumerConf NoWait 控制服务器响应机制，设置true为非阻塞模式，连接服务的时候，不会等待服务反馈成功的响应，就执行消费者，无法感知消费者是否创建成功，设置false，阻塞模式，会等待服务响应，成功才进行消费
type ConsumerConf struct {
	Name          string
	MaxRetryCount int64 `json:",default=3"`
	AutoAck       bool  `json:",default=false"`
	Exclusive     bool  `json:",default=false"`
	NoLocal       bool  `json:",default=false"`
	NoWait        bool  `json:",default=false"`
}

type QueueConf struct {
	Name       string
	Durable    bool `json:",default=true"`
	AutoDelete bool `json:",default=false"`
	Exclusive  bool `json:",default=false"`
	NoWait     bool `json:",default=false"`
}

type ExchangeConf struct {
	ExchangeName string
	Type         string `json:",options=direct|fanout|topic|headers"` // exchange type
	Durable      bool   `json:",default=true"`
	AutoDelete   bool   `json:",default=false"`
	Internal     bool   `json:",default=false"`
	NoWait       bool   `json:",default=false"`
	Queues       []QueueConf
}

// ConsumeHandler 允许客户端注入的消费逻辑
type ConsumeHandler interface {
	Consume(ctx context.Context, message []byte) error
}

type (
	Sender interface {
		Send(ctx context.Context, exchange string, routeKey string, msg []byte) error
		Close() error
	}

	RabbitMqSender struct {
		conn        *amqp.Connection
		channel     *amqp.Channel
		ContentType string
	}
)
