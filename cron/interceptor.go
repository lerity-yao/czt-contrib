package cron

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/hibiken/asynq"
	"github.com/zeromicro/go-zero/core/logc"
)

// RecoveryMiddleware 中间件
func RecoveryMiddleware(next asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logc.Errorf(ctx, "[CRON_PANIC] type: %s, panic: %v\n%s", t.Type(), r, debug.Stack())
				// 记录 panic 指标
				MetricServerPanicTotal.Inc(t.Type())
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
			logc.Errorf(ctx, "[CRON_ERROR] type: %s, err: %v, payload: %s", t.Type(), err, string(t.Payload()))
		}
		return err
	})
}

// PrometheusMiddleware 中间件
func PrometheusMiddleware(next asynq.Handler) asynq.Handler {
	return asynq.HandlerFunc(func(ctx context.Context, t *asynq.Task) error {
		taskType := t.Type()
		payloadSize := len(t.Payload())

		// 记录并发占用
		MetricServerActiveWorkers.Inc(taskType)
		defer MetricServerActiveWorkers.Dec(taskType)

		// 记录消费字节数
		MetricServerConsumeBytes.Add(float64(payloadSize), taskType)

		// 检查是否是重试执行
		retried, _ := asynq.GetRetryCount(ctx)
		if retried > 0 {
			MetricServerRetryTotal.Inc(taskType)
		}

		start := time.Now()
		err := next.ProcessTask(ctx, t)
		durationMs := time.Since(start).Milliseconds()

		// 记录耗时
		MetricServerConsumeDuration.Observe(durationMs, taskType)

		// 记录消费结果
		if err != nil {
			if isSkipRetry(err) {
				MetricServerSkipRetryTotal.Inc(taskType)
				MetricServerConsumeTotal.Inc(taskType, "skip_retry")
			} else {
				MetricServerConsumeTotal.Inc(taskType, "fail")
			}
		} else {
			MetricServerConsumeTotal.Inc(taskType, "success")
		}

		return err
	})
}

// isSkipRetry 检查是否是跳过重试的错误
func isSkipRetry(err error) bool {
	return errors.Is(err, asynq.SkipRetry)
}
