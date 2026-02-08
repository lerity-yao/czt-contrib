package cron

import "time"

func toDuration(v int64) time.Duration {
	if v <= 0 {
		return 0
	}
	return time.Duration(v) * time.Second
}
