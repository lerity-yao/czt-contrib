package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/zeromicro/go-zero/core/logc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

// Interceptor 拦截器定义
type Interceptor func(ctx context.Context, queueName string, message []byte, next func(context.Context, []byte) error) error

// Chain 拦截器链构造器
func Chain(interceptors ...Interceptor) Interceptor {
	return func(ctx context.Context, queueName string, message []byte, next func(context.Context, []byte) error) error {
		chained := next
		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor := interceptors[i]
			currentNext := chained
			chained = func(c context.Context, b []byte) error {
				return interceptor(c, queueName, b, currentNext)
			}
		}
		return chained(ctx, message)
	}
}

// --- 内置拦截器集 ---

// 1. Recovery 拦截器：捕获 Panic 并转化为 Error
func recoveryInterceptor(ctx context.Context, queueName string, body []byte, next func(context.Context, []byte) error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			logc.Errorf(ctx, "[RABBITMQ_PANIC] queue: %s, panic: %v\n%s", queueName, r, debug.Stack())
			metricListenerPanicTotal.Inc(queueName)
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return next(ctx, body)
}

// 2. Trace 拦截器：解析消息并注入链路
func traceInterceptor(ctx context.Context, queueName string, body []byte, next func(context.Context, []byte) error) error {
	var msgBody RabbitMsgBody
	if err := json.Unmarshal(body, &msgBody); err != nil {
		logc.Errorf(ctx, "[RABBITMQ_PARSE_ERROR] queue: %s, err: %v, payload: %s", queueName, err, string(body))
		metricListenerParseErrorTotal.Inc(queueName)
		return fmt.Errorf("failed to parse message: %w", err)
	}

	// 开启消费者 Span，传递解析后的业务消息
	childCtx, span := StartConsumerSpan(ctx, queueName, msgBody.Carrier)
	err := next(childCtx, msgBody.Msg)
	EndSpan(span, err)
	return err
}

// 3. Prometheus 拦截器：自动感知监控
func prometheusInterceptor(ctx context.Context, queueName string, body []byte, next func(context.Context, []byte) error) error {
	start := time.Now()

	// 记录消息大小
	metricListenerConsumeSize.Observe(int64(len(body)), queueName)

	err := next(ctx, body)

	durationMs := time.Since(start).Milliseconds()
	metricListenerConsumeDuration.Observe(durationMs, queueName)

	status := "success"
	if err != nil {
		status = "fail"
	}
	metricListenerConsumeTotal.Inc(queueName, status)
	return err
}

// 4. Logging 拦截器：统一异常记录
func loggingInterceptor(ctx context.Context, queueName string, body []byte, next func(context.Context, []byte) error) error {
	err := next(ctx, body)
	if err != nil {
		logc.Errorf(ctx, "[RABBITMQ_ERROR] queue: %s, err: %v, payload: %s", queueName, err, string(body))
	}
	return err
}

// ==================== Sender 拦截器 ====================

// SenderInterceptor Sender 拦截器定义
type SenderInterceptor func(ctx context.Context, exchange, routeKey string, msg []byte, next SenderFunc) error

// SenderFunc 发送函数签名
type SenderFunc func(ctx context.Context, msg []byte) error

// SenderChain Sender 拦截器链构造器
func SenderChain(interceptors ...SenderInterceptor) SenderInterceptor {
	return func(ctx context.Context, exchange, routeKey string, msg []byte, next SenderFunc) error {
		chained := next
		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor := interceptors[i]
			currentNext := chained
			chained = func(c context.Context, b []byte) error {
				return interceptor(c, exchange, routeKey, b, currentNext)
			}
		}
		return chained(ctx, msg)
	}
}

// --- Sender 内置拦截器 ---

// 1. Prometheus 拦截器：记录发送指标
func senderPrometheusInterceptor(ctx context.Context, exchange, routeKey string, msg []byte, next SenderFunc) error {
	start := time.Now()

	// 记录消息大小（原始业务消息）
	metricSenderSendSize.Observe(int64(len(msg)), exchange, routeKey)

	err := next(ctx, msg)

	durationMs := time.Since(start).Milliseconds()
	metricSenderSendDuration.Observe(durationMs, exchange, routeKey)

	status := "success"
	if err != nil {
		status = "fail"
	}
	metricSenderSendTotal.Inc(exchange, routeKey, status)
	return err
}

// 2. Logging 拦截器：记录发送日志
func senderLoggingInterceptor(ctx context.Context, exchange, routeKey string, msg []byte, next SenderFunc) error {
	err := next(ctx, msg)
	if err != nil {
		logc.Errorf(ctx, "[RABBITMQ_SEND_ERROR] exchange: %s, routeKey: %s, err: %v", exchange, routeKey, err)
	} else {
		logc.Infof(ctx, "[RABBITMQ_SEND_OK] exchange: %s, routeKey: %s, msg: %s", exchange, routeKey, string(msg))
	}
	return err
}

// 3. Trace 拦截器：注入链路并包装消息
func senderTraceInterceptor(ctx context.Context, exchange, routeKey string, msg []byte, next SenderFunc) error {
	// 开启生产者 Span
	ctx, span := StartProducerSpan(ctx, exchange, routeKey)
	defer func() {
		if r := recover(); r != nil {
			EndSpan(span, fmt.Errorf("panic: %v", r))
			panic(r)
		}
	}()

	// 注入 trace 上下文到 carrier
	carrier := &propagation.HeaderCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)

	msgBody := &RabbitMsgBody{
		Carrier: carrier,
		Msg:     msg,
	}

	wrappedMsg, err := json.Marshal(msgBody)
	if err != nil {
		EndSpan(span, err)
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	err = next(ctx, wrappedMsg)
	EndSpan(span, err)
	return err
}
