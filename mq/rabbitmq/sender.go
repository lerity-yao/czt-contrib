package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type (
	Sender interface {
		Send(ctx context.Context, exchange string, routeKey string, msg []byte) error
		Close() error
	}

	RabbitMqSender struct {
		conn           *amqp.Connection
		channel        *amqp.Channel
		ContentType    string
		rabbitConf     RabbitConf
		maxRetry       int
		reconnectMutex sync.Mutex
		closed         atomic.Bool // 标记是否已收到停止信号
	}
)

func MustNewSender(ctx context.Context, rabbitSenderConf RabbitSenderConf) Sender {
	logc.Infof(ctx, "rabbitmq sender: %v", rabbitSenderConf)
	s, err := NewSender(rabbitSenderConf)
	if err != nil {
		logx.Must(err)
	}
	return s
}

func NewSender(rabbitMqConf RabbitSenderConf) (Sender, error) {
	sender := &RabbitMqSender{
		ContentType: rabbitMqConf.ContentType,
		rabbitConf:  rabbitMqConf.RabbitConf,
		maxRetry:    10,
	}
	if err := sender.connect(); err != nil {
		return nil, err
	}

	// 注册优雅关闭钩子
	proc.AddShutdownListener(func() {
		logx.Info("Shutting down RabbitMQ sender...")
		sender.closed.Store(true) // 标记已收到停止信号，不再重连
		if err := sender.Close(); err != nil {
			logx.Errorf("Failed to close RabbitMQ sender: %v", err)
		} else {
			logx.Info("RabbitMQ sender shut down gracefully")
		}
	})

	return sender, nil
}

func (q *RabbitMqSender) connect() error {
	var err error
	maxRetry := 0
	for maxRetry < q.maxRetry {
		q.conn, err = amqp.DialConfig(getRabbitURL(q.rabbitConf), amqp.Config{
			Heartbeat: 30 * time.Second,
		})
		if err == nil {
			logx.Infof("Connected to RabbitMQ")
			break
		}
		maxRetry++
		logx.Errorf("Failed to connect to RabbitMQ: %v. Retrying(%d/%d)", err, maxRetry, q.maxRetry)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		logx.Errorf("Failed to connect to RabbitMQ after %d retries", q.maxRetry)
		if q.conn != nil {
			q.conn.Close()
			q.conn = nil
		}
		return fmt.Errorf("failed to connect rabbitmq, error: %v", err)
	}

	q.handleConnectionClose()

	maxRetry = 0
	for maxRetry < q.maxRetry {
		q.channel, err = q.conn.Channel()
		if err == nil {
			logx.Infof("Channel created successfully")
			break
		}
		maxRetry++
		logx.Errorf("Failed to open a channel: %v. Retrying(%d/%d)", err, maxRetry, q.maxRetry)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		logx.Errorf("Failed to open a channel after %d retries", q.maxRetry)
		if q.conn != nil {
			q.conn.Close()
			q.conn = nil
		}
		return fmt.Errorf("failed to open a channel, error: %v", err)
	}

	q.handleChannelClose()

	return nil
}

func (q *RabbitMqSender) handleConnectionClose() {
	connCloseChan := q.conn.NotifyClose(make(chan *amqp.Error))

	go func() {
		for err := range connCloseChan {
			if q.closed.Load() {
				logx.Info("Received shutdown signal, skip reconnect on connection close")
				return
			}
			logx.Errorf("Sender connection closed: %v", err)
			metricSenderDisconnectTotal.Inc()
			_ = q.reconnect()
		}
	}()
}

func (q *RabbitMqSender) handleChannelClose() {
	chanCloseChan := q.channel.NotifyClose(make(chan *amqp.Error))

	go func() {
		for err := range chanCloseChan {
			if q.closed.Load() {
				logx.Info("Received shutdown signal, skip reconnect on channel close")
				return
			}
			logx.Errorf("Sender channel closed: %v", err)
			metricSenderDisconnectTotal.Inc()
			_ = q.reconnect()
		}
	}()
}

func (q *RabbitMqSender) reconnect() error {
	// 收到停止信号后不再重连
	if q.closed.Load() {
		logx.Info("Received shutdown signal, skip reconnect")
		return errors.New("sender is closing, skip reconnect")
	}

	logx.Info("Attempting to reconnect...")
	q.reconnectMutex.Lock()
	defer q.reconnectMutex.Unlock()

	// 加锁后再次检查，避免重复重连
	if q.closed.Load() {
		logx.Info("Received shutdown signal after lock, skip reconnect")
		return errors.New("sender is closing, skip reconnect")
	}
	if q.conn != nil && !q.conn.IsClosed() && q.channel != nil {
		logx.Info("Already reconnected, skip")
		return nil
	}

	if q.channel != nil {
		_ = q.channel.Close()
		q.channel = nil
	}
	if q.conn != nil {
		_ = q.conn.Close()
		q.conn = nil
	}

	if err := q.connect(); err != nil {
		logx.Errorf("Reconnect failed: %v", err)
		return err
	}

	metricSenderReconnectTotal.Inc()
	logx.Info("Reconnect success")
	return nil
}

func (q *RabbitMqSender) Send(ctx context.Context, exchange string, routeKey string, msg []byte) error {
	start := time.Now()

	// 检查连接和通道状态，如果已关闭尝试重连
	if q.conn == nil || q.conn.IsClosed() || q.channel == nil {
		if err := q.reconnect(); err != nil {
			metricSenderSendTotal.Inc(exchange, routeKey, "fail")
			metricSenderSendDuration.Observe(time.Since(start).Milliseconds(), exchange, routeKey)
			return errors.New("connection closed and reconnect failed")
		}
	}

	// 记录消息大小
	metricSenderSendSize.Observe(int64(len(msg)), exchange, routeKey)

	// 开启生产者 Span
	_, span := StartProducerSpan(ctx, exchange, routeKey)
	defer func() {
		if err := recover(); err != nil {
			EndSpan(span, fmt.Errorf("panic: %v", err))
			panic(err)
		}
	}()

	// 注入 trace 上下文到 carrier
	carrier := &propagation.HeaderCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	msgBody := &RabbitMsgBody{
		Carrier: carrier,
		Msg:     msg,
	}

	msgBodyMap, err := json.Marshal(msgBody)
	if err != nil {
		EndSpan(span, err)
		metricSenderSendTotal.Inc(exchange, routeKey, "fail")
		metricSenderSendDuration.Observe(time.Since(start).Milliseconds(), exchange, routeKey)
		return err
	}

	err = q.channel.PublishWithContext(
		ctx,
		exchange,
		routeKey,
		false,
		false,
		amqp.Publishing{
			ContentType: q.ContentType,
			Body:        msgBodyMap,
		},
	)

	if err != nil {
		logc.Infof(ctx, "Failed to publish a message, error: %v", err)
		EndSpan(span, err)
		metricSenderSendTotal.Inc(exchange, routeKey, "fail")
		metricSenderSendDuration.Observe(time.Since(start).Milliseconds(), exchange, routeKey)
		return err
	}

	logc.Infof(ctx, "Successfully publish a message, message: %v", string(msg))
	EndSpan(span, nil)
	metricSenderSendTotal.Inc(exchange, routeKey, "success")
	metricSenderSendDuration.Observe(time.Since(start).Milliseconds(), exchange, routeKey)
	return nil
}

func (q *RabbitMqSender) Close() error {
	logx.Info("Closing RabbitMQ sender...")
	q.closed.Store(true) // 标记已关闭，防止触发重连

	// 先关闭 channel
	if q.channel != nil {
		if err := q.channel.Close(); err != nil {
			logx.Errorf("Failed to close channel: %v", err)
		} else {
			logx.Info("Channel closed")
		}
		q.channel = nil
	}

	// 再关闭 connection
	if q.conn != nil {
		if err := q.conn.Close(); err != nil {
			logx.Errorf("Failed to close connection: %v", err)
		} else {
			logx.Info("Connection closed")
		}
		q.conn = nil
	}

	logx.Info("RabbitMQ sender closed")
	return nil
}
