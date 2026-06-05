package cron

import (
	"context"
	"math"
	"time"

	"github.com/hibiken/asynq"
)

func toDuration(v int64) time.Duration {
	if v <= 0 {
		return 0
	}
	return time.Duration(v) * time.Second
}

// ExponentialRetryDelay 指数退避策略，延迟为 (2^n - 1) 秒。
// 重试序列：1s, 3s, 7s, 15s, 31s, 63s, 127s, ...
// n 为上次重试次数
func ExponentialRetryDelay(n int, _ error, _ *asynq.Task) time.Duration {
	return time.Duration(math.Pow(2, float64(n+1))-1) * time.Second
}

// GetTaskID 从 context 中提取任务 ID。
// 任务 ID 在重试过程中保持不变。
func GetTaskID(ctx context.Context) (string, bool) {
	return asynq.GetTaskID(ctx)
}

// GetRetryCount 从 context 中提取当前重试次数。
// 返回 0 表示首次执行。
func GetRetryCount(ctx context.Context) (int, bool) {
	return asynq.GetRetryCount(ctx)
}

// GetMaxRetry 从 context 中提取最大重试次数。
func GetMaxRetry(ctx context.Context) (int, bool) {
	return asynq.GetMaxRetry(ctx)
}

// GetQueueName 从 context 中提取任务所在队列名。
func GetQueueName(ctx context.Context) (string, bool) {
	return asynq.GetQueueName(ctx)
}
