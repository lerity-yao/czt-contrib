package cron

import (
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
func ExponentialRetryDelay(n int, _ error, _ *asynq.Task) time.Duration {
	return time.Duration(math.Pow(2, float64(n))-1) * time.Second
}
