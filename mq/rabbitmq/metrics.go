package rabbitmq

import "github.com/zeromicro/go-zero/core/metric"

// ==================== Sender 指标 ====================

var (
	// 发送总数 (exchange, route_key, status)
	metricSenderSendTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Name:   "mq_sender_send_total",
		Help:   "RabbitMQ 消息发送总数",
		Labels: []string{"exchange", "route_key", "status"},
	})

	// 发送耗时 (exchange, route_key)
	metricSenderSendDuration = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Name:    "mq_sender_send_duration_ms",
		Help:    "RabbitMQ 消息发送耗时(ms)",
		Labels:  []string{"exchange", "route_key"},
		Buckets: []float64{1, 2, 5, 10, 25, 50, 100, 250, 500, 1000},
	})

	// 发送消息大小 (exchange, route_key)
	metricSenderSendSize = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Name:    "mq_sender_send_size_bytes",
		Help:    "RabbitMQ 消息发送大小(bytes)",
		Labels:  []string{"exchange", "route_key"},
		Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000, 1000000},
	})

	// 重连次数
	metricSenderReconnectTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Name:   "mq_sender_reconnect_total",
		Help:   "RabbitMQ Sender 重连次数",
		Labels: []string{},
	})

	// 掉线次数
	metricSenderDisconnectTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Name:   "mq_sender_disconnect_total",
		Help:   "RabbitMQ Sender 掉线次数",
		Labels: []string{},
	})
)

// ==================== Listener 指标 ====================

var (
	// 消费总数 (queue, status)
	metricListenerConsumeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Name:   "mq_listener_consume_total",
		Help:   "RabbitMQ 消息消费总数",
		Labels: []string{"queue", "status"},
	})

	// 消费耗时 (queue)
	metricListenerConsumeDuration = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Name:    "mq_listener_consume_duration_ms",
		Help:    "RabbitMQ 消息消费耗时(ms)",
		Labels:  []string{"queue"},
		Buckets: []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
	})

	// 消费消息大小 (queue)
	metricListenerConsumeSize = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Name:    "mq_listener_consume_size_bytes",
		Help:    "RabbitMQ 消息消费大小(bytes)",
		Labels:  []string{"queue"},
		Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000, 500000, 1000000},
	})

	// 当前处理中的消息数 (queue)
	metricListenerInFlight = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Name:   "mq_listener_in_flight",
		Help:   "RabbitMQ 当前正在处理的消息数",
		Labels: []string{"queue"},
	})

	// 解析失败次数 (queue)
	metricListenerParseErrorTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Name:   "mq_listener_parse_error_total",
		Help:   "RabbitMQ 消息解析失败次数",
		Labels: []string{"queue"},
	})

	// Panic 次数 (queue)
	metricListenerPanicTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Name:   "mq_listener_panic_total",
		Help:   "RabbitMQ 消费 Panic 次数",
		Labels: []string{"queue"},
	})

	// ACK 计数 (queue, type)
	metricListenerAckTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Name:   "mq_listener_ack_total",
		Help:   "RabbitMQ ACK/Reject 计数",
		Labels: []string{"queue", "type"},
	})

	// 重连次数
	metricListenerReconnectTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Name:   "mq_listener_reconnect_total",
		Help:   "RabbitMQ Listener 重连次数",
		Labels: []string{},
	})

	// 掉线次数
	metricListenerDisconnectTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Name:   "mq_listener_disconnect_total",
		Help:   "RabbitMQ Listener 掉线次数",
		Labels: []string{},
	})
)
