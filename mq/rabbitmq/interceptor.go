package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/zeromicro/go-zero/core/logc"
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
			logc.Errorf(ctx, "[MQ_PANIC] queue: %s, panic: %v\n%s", queueName, r, debug.Stack())
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
		logc.Errorf(ctx, "[MQ_PARSE_ERROR] queue: %s, err: %v, payload: %s", queueName, err, string(body))
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
		logc.Errorf(ctx, "[MQ_ERROR] queue: %s, err: %v, payload: %s", queueName, err, string(body))
	}
	return err
}
