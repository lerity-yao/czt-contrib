package rabbitmq

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/metric"
	"github.com/zeromicro/go-zero/core/trace"
	"go.opentelemetry.io/otel"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// Interceptor 拦截器定义
type Interceptor func(ctx context.Context, queueName string, message []byte, next func(context.Context, []byte) error) error

// 指标定义
var (
	metricConsumeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Name:   "mq_consume_total",
		Help:   "RabbitMQ 消费总数统计",
		Labels: []string{"queue", "status"},
	})
	metricConsumeDuration = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Name:    "mq_consume_duration_ms",
		Help:    "RabbitMQ 消费耗时统计(ms)",
		Labels:  []string{"queue"},
		Buckets: []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
	})
)

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
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	return next(ctx, body)
}

// 2. Trace 拦截器：自动感知链路并注入
func traceInterceptor(ctx context.Context, queueName string, body []byte, next func(context.Context, []byte) error) error {
	var msgBody RabbitMsgBody
	if err := json.Unmarshal(body, &msgBody); err != nil {
		return next(ctx, body)
	}

	propagator := otel.GetTextMapPropagator()
	extractedCtx := propagator.Extract(ctx, msgBody.Carrier)

	tracer := otel.GetTracerProvider().Tracer(trace.TraceName)
	childCtx, span := tracer.Start(extractedCtx,
		fmt.Sprintf("mq-consume-%s", queueName),
		oteltrace.WithSpanKind(oteltrace.SpanKindConsumer),
	)
	defer span.End()

	return next(childCtx, body)
}

// 3. Prometheus 拦截器：自动感知监控
func prometheusInterceptor(ctx context.Context, queueName string, body []byte, next func(context.Context, []byte) error) error {
	start := time.Now()
	err := next(ctx, body)

	durationMs := time.Since(start).Milliseconds()
	metricConsumeDuration.Observe(durationMs, queueName)

	status := "success"
	if err != nil {
		status = "fail"
	}
	metricConsumeTotal.Inc(queueName, status)
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
