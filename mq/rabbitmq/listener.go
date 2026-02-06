package rabbitmq

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/queue"
)

// MustNewListener rabbitmq消费者服务端
func MustNewListener(ctx context.Context, rabbitListenerConf RabbitListenerConf, handler ConsumeHandler) queue.MessageQueue {
	defaultInterceptor := Chain(
		recoveryInterceptor,
		traceInterceptor,
		prometheusInterceptor,
		loggingInterceptor,
	)

	currentCtx, cancel := context.WithCancel(ctx)

	listener := &RabbitListener{
		queues:      rabbitListenerConf,
		handler:     handler,
		forever:     make(chan bool),
		rootCtx:     ctx,
		currentCtx:  currentCtx,
		cancel:      cancel,
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
	return nil
}

func (q *RabbitListener) handleConnectionClose() {
	connCloseChan := q.conn.NotifyClose(make(chan *amqp.Error))

	go func() {
		for err := range connCloseChan {
			logx.Errorf("Connection closed: %v", err)
			q.reconnect()
		}
	}()
}

func (q *RabbitListener) reconnect() {
	logx.Info("Attempting to reconnect...")
	q.reconnectMutex.Lock()
	defer q.reconnectMutex.Unlock()

	if q.cancel != nil {
		q.cancel()
	}
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

	// 关键：重连时重新派生上下文并保存副本
	q.currentCtx, q.cancel = context.WithCancel(q.rootCtx)

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
		time.AfterFunc(5*time.Second, q.reconnect)
		return
	}

	q.internalStart(q.currentCtx)
}

func (q *RabbitListener) parseMessage(listenerConsumer ConsumerConf, message amqp.Delivery) (*RabbitMsgBody, error) {

	var msgBody = new(RabbitMsgBody)
	err := json.Unmarshal(message.Body, msgBody)
	if err != nil {
		logx.Errorf("Failed to parse RabbitMQ message payload, delivery: %v, error: %v", message, err)
		if listenerConsumer.AutoAck == false {
			// 紧确认消费当前消息，因为此消息当前消费者无法解析，同队列的其他消费者肯定也无法解析，需要确认消费掉，不然一直循环消费
			message.Ack(false)
		}
		return nil, err
	}
	return msgBody, nil
}

func (q *RabbitListener) requeueMessage(ctx context.Context, listenerConsumer ConsumerConf, message amqp.Delivery, retryCount int64) {
	if message.Headers == nil {
		message.Headers = make(amqp.Table)
	}
	message.Headers["x-retry-count"] = retryCount + 1
	// 手动重新发布消息（确保 headers 被保留）
	err := q.channel.Publish(
		"",
		listenerConsumer.Name,
		false,
		false,
		amqp.Publishing{
			Headers:     message.Headers,
			ContentType: q.queues.ContentType,
			Body:        message.Body,
		},
	)
	if err != nil {
		logc.Errorf(ctx, "Failed requeue message: %v,  err: %v", string(message.Body), err)
	}

	if err == nil {
		logc.Infof(ctx, "Successfully requeue message : %v", string(message.Body))
	}

	// 确认原消息（避免重复消费）
	if !listenerConsumer.AutoAck {
		message.Ack(false)
	}
}
func (q *RabbitListener) processMessage(listenerConsumer ConsumerConf, message amqp.Delivery, runCtx context.Context) {
	// 激进拒绝：使用当前协程分配到的 context 快照
	select {
	case <-runCtx.Done():
		_ = message.Reject(true)
		return
	default:
	}

	q.taskWg.Add(1)
	defer q.taskWg.Done()

	handleLogic := func(ctx context.Context, rawBody []byte) error {
		// 业务逻辑内的 context 二次检查
		select {
		case <-ctx.Done():
			return message.Reject(true)
		default:
		}

		msgBody, err := q.parseMessage(listenerConsumer, message)
		if err != nil {
			return err
		}

		var retryCount int64
		if val, ok := message.Headers["x-retry-count"]; ok {
			switch v := val.(type) {
			case int64:
				retryCount = v
			case int32:
				retryCount = int64(v)
			case int:
				retryCount = int64(v)
			}
		}

		if retryCount > listenerConsumer.MaxRetryCount {
			if !listenerConsumer.AutoAck {
				_ = message.Ack(false)
			}
			return nil
		}

		err = q.handler.Consume(ctx, msgBody.Msg)
		if err != nil {
			time.Sleep(time.Millisecond * 100)
			q.requeueMessage(ctx, listenerConsumer, message, retryCount)
			return err
		}

		if !listenerConsumer.AutoAck {
			_ = message.Ack(false)
		}
		return nil
	}

	_ = q.interceptor(runCtx, listenerConsumer.Name, message.Body, handleLogic)
}

func (q *RabbitListener) internalStart(runCtx context.Context) {
	for i := range q.queues.ListenerQueues {
		lQueue := q.queues.ListenerQueues[i]
		q.listenerWg.Add(1)
		go func(lConsumer ConsumerConf, ctx context.Context) {
			defer q.listenerWg.Done()
			queueMessages, err := q.channel.Consume(lConsumer.Name, "", lConsumer.AutoAck, false, false, false, nil)
			if err != nil {
				logx.Errorf("Failed to consume %s: %v", lConsumer.Name, err)
				return
			}
			for {
				select {
				case <-ctx.Done(): // 精准捕获该监听协程对应的信号
					logx.Infof("Exit consumer loop for: %s", lConsumer.Name)
					return
				case message, ok := <-queueMessages:
					if !ok {
						return
					}
					q.processMessage(lConsumer, message, ctx)
				}
			}
		}(lQueue, runCtx)
	}
}

func (q *RabbitListener) Start() {
	q.internalStart(q.currentCtx)
	<-q.forever
}

func (q *RabbitListener) Stop() {
	logx.Info("RabbitMQ Listener is shutting down...")
	if q.cancel != nil {
		q.cancel()
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

	if q.channel != nil {

		_ = q.channel.Close()
	}
	if q.conn != nil {
		_ = q.conn.Close()
	}
	close(q.forever)
}
