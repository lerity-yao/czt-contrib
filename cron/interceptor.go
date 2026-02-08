package cron

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/metric"
)

var (
	metricConsumeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Name:   "cron_consume_total",
		Help:   "消费总数统计",
		Labels: []string{"task_type", "status"},
	})

	metricConsumeDuration = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Name:    "cron_consume_duration_ms",
		Help:    "消费耗时统计(ms)",
		Labels:  []string{"task_type"},
		Buckets: []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
	})

	metricActiveWorkers = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Name:   "cron_active_workers",
		Help:   "当前正在执行的任务并发数",
		Labels: []string{"task_type"},
	})
)

// RecoveryMiddleware 中间件
func RecoveryMiddleware(next asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logc.Errorf(ctx, "[ASYNQ_PANIC] type: %s, panic: %v\n%s", t.Type(), r, debug.Stack())
				// panic 不重试
				err = fmt.Errorf("%w: panic occurred: %v", asynq.SkipRetry, r)
			}
		}()
		return next.ProcessTask(ctx, t)
	})
}

// TraceMiddleware 中间件 (从 Header 提取 Carrier)
func TraceMiddleware(next asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) (err error) {
		childCtx, span := StartConsumerSpan(ctx, t)
		defer func() {
			EndSpan(span, err)
		}()
		return next.ProcessTask(childCtx, t)
	})
}

// LoggingMiddleware 中间件
func LoggingMiddleware(next asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		err := next.ProcessTask(ctx, t)
		if err != nil {
			logc.Errorf(ctx, "[ASYNQ_ERROR] type: %s, err: %v, payload: %s", t.Type(), err, string(t.Payload()))
		}
		return err
	})
}

// PrometheusMiddleware 中间件
func PrometheusMiddleware(next asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		// --- [核心1] 记录并发占用 (开始+1, 结束-1) ---
		metricActiveWorkers.Inc(t.Type())
		defer metricActiveWorkers.Dec(t.Type())

		start := time.Now()
		err := next.ProcessTask(ctx, t)

		durationMs := time.Since(start).Milliseconds()
		metricConsumeDuration.Observe(durationMs, t.Type())

		status := "success"
		if err != nil {
			status = "fail"
		}
		metricConsumeTotal.Inc(t.Type(), status)
		return err
	})
}
