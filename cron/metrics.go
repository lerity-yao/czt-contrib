package cron

import (
	"log"

	"github.com/hibiken/asynq"
	"github.com/prometheus/client_golang/prometheus"
)

const metricsNamespace = "cron"

// QueueMetricsCollector 带队列过滤的指标采集器
// 只采集 allowedQueues 中配置的队列指标，解决多服务共用 Redis 时指标混杂问题
type QueueMetricsCollector struct {
	inspector     *asynq.Inspector
	allowedQueues map[string]struct{} // 允许采集的队列白名单
}

// Descriptors - 与官方 collector 保持一致的指标定义
var (
	tasksQueuedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricsNamespace, "", "tasks_enqueued_total"),
		"Number of tasks enqueued; broken down by queue and state.",
		[]string{"queue", "state"}, nil,
	)

	queueSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricsNamespace, "", "queue_size"),
		"Number of tasks in a queue",
		[]string{"queue"}, nil,
	)

	queueLatencyDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricsNamespace, "", "queue_latency_seconds"),
		"Number of seconds the oldest pending task is waiting in pending state to be processed.",
		[]string{"queue"}, nil,
	)

	queueMemUsgDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricsNamespace, "", "queue_memory_usage_approx_bytes"),
		"Number of memory used by a given queue (approximated number by sampling).",
		[]string{"queue"}, nil,
	)

	tasksProcessedTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricsNamespace, "", "tasks_processed_total"),
		"Number of tasks processed (both succeeded and failed); broken down by queue",
		[]string{"queue"}, nil,
	)

	tasksFailedTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricsNamespace, "", "tasks_failed_total"),
		"Number of tasks failed; broken down by queue",
		[]string{"queue"}, nil,
	)

	pausedQueuesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricsNamespace, "", "queue_paused_total"),
		"Number of queues paused",
		[]string{"queue"}, nil,
	)
)

// NewQueueMetricsCollector 创建带队列过滤的指标采集器
//
// 参数:
//   - inspector: asynq Inspector 实例
//   - allowedQueues: 允许采集的队列名列表，只采集这些队列的指标
//
// 设计说明:
//   - 官方 metrics.NewQueueMetricsCollector 会采集 Redis 中所有队列的指标
//   - 当多个服务共用同一个 Redis 实例时，会导致服务 A 的 /metrics 端点包含服务 B 的队列指标
//   - 本实现通过白名单过滤，只采集当前服务配置的队列，实现指标隔离
func NewQueueMetricsCollector(inspector *asynq.Inspector, allowedQueues []string) *QueueMetricsCollector {
	queueSet := make(map[string]struct{}, len(allowedQueues))
	for _, q := range allowedQueues {
		queueSet[q] = struct{}{}
	}
	return &QueueMetricsCollector{
		inspector:     inspector,
		allowedQueues: queueSet,
	}
}

func (c *QueueMetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	prometheus.DescribeByCollect(c, ch)
}

func (c *QueueMetricsCollector) Collect(ch chan<- prometheus.Metric) {
	queueInfos, err := c.collectQueueInfo()
	if err != nil {
		log.Printf("[ASYNQ_METRICS] Failed to collect metrics: %v", err)
		return
	}

	for _, info := range queueInfos {
		// 各状态任务数量
		ch <- prometheus.MustNewConstMetric(tasksQueuedDesc, prometheus.GaugeValue, float64(info.Active), info.Queue, "active")
		ch <- prometheus.MustNewConstMetric(tasksQueuedDesc, prometheus.GaugeValue, float64(info.Pending), info.Queue, "pending")
		ch <- prometheus.MustNewConstMetric(tasksQueuedDesc, prometheus.GaugeValue, float64(info.Scheduled), info.Queue, "scheduled")
		ch <- prometheus.MustNewConstMetric(tasksQueuedDesc, prometheus.GaugeValue, float64(info.Retry), info.Queue, "retry")
		ch <- prometheus.MustNewConstMetric(tasksQueuedDesc, prometheus.GaugeValue, float64(info.Archived), info.Queue, "archived")
		ch <- prometheus.MustNewConstMetric(tasksQueuedDesc, prometheus.GaugeValue, float64(info.Completed), info.Queue, "completed")

		// 队列维度指标
		ch <- prometheus.MustNewConstMetric(queueSizeDesc, prometheus.GaugeValue, float64(info.Size), info.Queue)
		ch <- prometheus.MustNewConstMetric(queueLatencyDesc, prometheus.GaugeValue, info.Latency.Seconds(), info.Queue)
		ch <- prometheus.MustNewConstMetric(queueMemUsgDesc, prometheus.GaugeValue, float64(info.MemoryUsage), info.Queue)
		ch <- prometheus.MustNewConstMetric(tasksProcessedTotalDesc, prometheus.CounterValue, float64(info.ProcessedTotal), info.Queue)
		ch <- prometheus.MustNewConstMetric(tasksFailedTotalDesc, prometheus.CounterValue, float64(info.FailedTotal), info.Queue)

		// 队列暂停状态
		pausedValue := 0
		if info.Paused {
			pausedValue = 1
		}
		ch <- prometheus.MustNewConstMetric(pausedQueuesDesc, prometheus.GaugeValue, float64(pausedValue), info.Queue)
	}
}

// collectQueueInfo 采集队列信息，只返回白名单中的队列
func (c *QueueMetricsCollector) collectQueueInfo() ([]*asynq.QueueInfo, error) {
	// 如果白名单为空，不采集任何队列（避免误采集全量）
	if len(c.allowedQueues) == 0 {
		return nil, nil
	}

	var infos []*asynq.QueueInfo
	for queueName := range c.allowedQueues {
		qinfo, err := c.inspector.GetQueueInfo(queueName)
		if err != nil {
			// 队列不存在时跳过，不中断采集
			log.Printf("[ASYNQ_METRICS] Failed to get queue info for %s: %v", queueName, err)
			continue
		}
		infos = append(infos, qinfo)
	}
	return infos, nil
}
