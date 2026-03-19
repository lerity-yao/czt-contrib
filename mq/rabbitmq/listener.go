package rabbitmq

import (
	"context"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/queue"
)

// MustNewListener rabbitmq消费者服务端
func MustNewListener(rabbitListenerConf RabbitListenerConf, handler ConsumeHandler) queue.MessageQueue {
	defaultInterceptor := Chain(
		recoveryInterceptor,
		prometheusInterceptor,
		loggingInterceptor,
		traceInterceptor,
	)

	listener := &RabbitListener{
		queues:      rabbitListenerConf,
		handler:     handler,
		forever:     make(chan bool),
		taskWg:      sync.WaitGroup{},
		listenerWg:  sync.WaitGroup{},
		maxRetry:    10,
		interceptor: defaultInterceptor,
	}
	err := listener.connect()
	logx.Must(err)
	return listener
}

func (q *RabbitListener) connect() error {
	var err error
	maxRetry := 0
	for maxRetry < q.maxRetry {
		q.conn, err = amqp.DialConfig(getRabbitURL(q.queues.RabbitConf), amqp.Config{
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
		logx.Errorf("Failed to connect to RabbitMQ, Reconnect...")
		if q.conn != nil {
			q.conn.Close()
			q.conn = nil
		}
		return err
	}

	q.handleConnectionClose()

	maxRetry = 0
	for maxRetry < q.maxRetry {
		q.channel, err = q.conn.Channel()
		if err == nil {
			logx.Infof("Channel created successfully")
			err = q.channel.Qos(
				q.queues.ChannelQos.PrefetchCount,
				q.queues.ChannelQos.PrefetchSize,
				q.queues.ChannelQos.Global,
			)
			if err != nil {
				logx.Errorf("Failed to set QoS: %v. Retrying(%d/%d)", err, maxRetry, q.maxRetry)
				maxRetry++
				continue
			}
			logx.Infof("Successfully to set QoS prefetchCount: %d, prefetchSize: %d, global: %t",
				q.queues.ChannelQos.PrefetchCount, q.queues.ChannelQos.PrefetchSize, q.queues.ChannelQos.Global)
			break
		}
		maxRetry++
		logx.Errorf("Failed to open a channel: %v. Retrying(%d/%d)", err, maxRetry, q.maxRetry)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		logx.Errorf("Failed to open a channel, Reconnect...")
		if q.conn != nil {
			q.conn.Close()
			q.conn = nil
		}
		return err
	}

	q.handleChannelClose()

	return nil
}

func (q *RabbitListener) handleConnectionClose() {
	connCloseChan := q.conn.NotifyClose(make(chan *amqp.Error))

	go func() {
		for err := range connCloseChan {
			if q.closed.Load() {
				logx.Info("Received shutdown signal, skip reconnect on connection close")
				return
			}
			logx.Errorf("Connection closed: %v", err)
			metricListenerDisconnectTotal.Inc()
			q.reconnect()
		}
	}()
}

func (q *RabbitListener) handleChannelClose() {
	chanCloseChan := q.channel.NotifyClose(make(chan *amqp.Error))

	go func() {
		for err := range chanCloseChan {
			if q.closed.Load() {
				logx.Info("Received shutdown signal, skip reconnect on channel close")
				return
			}
			logx.Errorf("Channel closed: %v", err)
			metricListenerDisconnectTotal.Inc()
			q.reconnect()
		}
	}()
}

func (q *RabbitListener) reconnect() {
	// 收到停止信号后不再重连
	if q.closed.Load() {
		logx.Info("Received shutdown signal, skip reconnect")
		return
	}

	logx.Info("Attempting to reconnect...")
	q.reconnectMutex.Lock()
	defer q.reconnectMutex.Unlock()

	// 加锁后再次检查，避免重复重连
	if q.closed.Load() {
		logx.Info("Received shutdown signal after lock, skip reconnect")
		return
	}
	if q.conn != nil && !q.conn.IsClosed() && q.channel != nil {
		logx.Info("Already reconnected, skip")
		return
	}

	// 等待旧 goroutine 自然退出（channel 关闭时自动退出）
	done := make(chan struct{})
	go func() {
		q.listenerWg.Wait()
		q.taskWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		logx.Info("All old routines exited, proceeding with reconnect.")
	case <-time.After(10 * time.Second):
		logx.Errorf("Wait for old routines timeout during reconnect, forcing reconnect.")
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
		return
	}

	metricListenerReconnectTotal.Inc()
	q.internalStart()
}

func (q *RabbitListener) processMessage(listenerConsumer ConsumerConf, message amqp.Delivery) {
	// 激进拒绝：检查停止信号
	if q.closed.Load() {
		_ = message.Reject(true)
		metricListenerAckTotal.Inc(listenerConsumer.Name, "reject")
		return
	}

	q.taskWg.Add(1)
	metricListenerInFlight.Inc(listenerConsumer.Name)
	defer func() {
		metricListenerInFlight.Dec(listenerConsumer.Name)
		q.taskWg.Done()
	}()

	handleLogic := func(ctx context.Context, rawBody []byte) error {
		// 业务逻辑内的二次检查
		if q.closed.Load() {
			return context.Canceled
		}
		// rawBody 已经是 traceInterceptor 解析后的业务消息
		return q.handler.Consume(ctx, rawBody)
	}

	_ = q.interceptor(context.Background(), listenerConsumer.Name, message.Body, handleLogic)

	// 统一处理消息确认
	if !listenerConsumer.AutoAck {
		if q.closed.Load() {
			_ = message.Reject(true) // 停止信号 → 重入队列
			metricListenerAckTotal.Inc(listenerConsumer.Name, "reject")
		} else {
			_ = message.Ack(false) // 其他情况（成功或失败）→ 确认消费
			metricListenerAckTotal.Inc(listenerConsumer.Name, "ack")
		}
	}
}

func (q *RabbitListener) internalStart() {
	for i := range q.queues.ListenerQueues {
		lQueue := q.queues.ListenerQueues[i]
		q.listenerWg.Add(1)
		go func(lConsumer ConsumerConf) {
			defer q.listenerWg.Done()
			queueMessages, err := q.channel.Consume(lConsumer.Name, "", lConsumer.AutoAck, false, false, false, nil)
			if err != nil {
				logx.Errorf("Failed to consume %s: %v", lConsumer.Name, err)
				return
			}
			for message := range queueMessages {
				if q.closed.Load() {
					logx.Infof("Exit consumer loop for: %s", lConsumer.Name)
					return
				}
				q.processMessage(lConsumer, message)
			}
		}(lQueue)
	}
}

func (q *RabbitListener) Start() {
	q.internalStart()
	<-q.forever
}

func (q *RabbitListener) Stop() {
	logx.Info("RabbitMQ Listener is shutting down...")
	q.closed.Store(true) // 标记已收到停止信号

	// 关闭 channel 让消费者 goroutine 自然退出
	if q.channel != nil {
		_ = q.channel.Close()
	}

	q.listenerWg.Wait() // 先停水龙头

	waitDone := make(chan struct{})
	go func() {
		q.taskWg.Wait() // 再排空水池
		close(waitDone)
	}()

	select {
	case <-waitDone:
		logx.Info("All processing tasks finished.")
	case <-time.After(time.Second * 10):
		logx.Error("Shutdown timeout.")
	}

	if q.conn != nil {
		_ = q.conn.Close()
	}
	close(q.forever)
}
