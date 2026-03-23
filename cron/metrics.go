package cron

import (
	"log"

	"github.com/hibiken/asynq"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zeromicro/go-zero/core/metric"
)

const metricsNamespace = "cron"

// ==================== Server 端指标（中间件采集，零侵入） ====================

const serverSubsystem = "server"

var (
	// MetricServerConsumeTotal 消费计数（成功/失败/skip_retry）
	MetricServerConsumeTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: metricsNamespace,
		Subsystem: serverSubsystem,
		Name:      "consume_total",
		Help:      "消费计数",
		Labels:    []string{"task_type", "status"},
	})

	// MetricServerConsumeDuration 消费耗时
	MetricServerConsumeDuration = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: metricsNamespace,
		Subsystem: serverSubsystem,
		Name:      "consume_duration_ms",
		Help:      "消费耗时统计(ms)",
		Labels:    []string{"task_type"},
		Buckets:   []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
	})

	// MetricServerConsumeBytes 消费字节数
	MetricServerConsumeBytes = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: metricsNamespace,
		Subsystem: serverSubsystem,
		Name:      "consume_bytes",
		Help:      "消费字节数",
		Labels:    []string{"task_type"},
	})

	// MetricServerActiveWorkers 当前并发数
	MetricServerActiveWorkers = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Namespace: metricsNamespace,
		Subsystem: serverSubsystem,
		Name:      "active_workers",
		Help:      "当前正在执行的任务并发数",
		Labels:    []string{"task_type"},
	})

	// MetricServerRetryTotal 重试次数（非首次执行）
	MetricServerRetryTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: metricsNamespace,
		Subsystem: serverSubsystem,
		Name:      "retry_total",
		Help:      "重试执行次数（非首次执行）",
		Labels:    []string{"task_type"},
	})

	// MetricServerSkipRetryTotal 跳过重试次数（主动放弃）
	MetricServerSkipRetryTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: metricsNamespace,
		Subsystem: serverSubsystem,
		Name:      "skip_retry_total",
		Help:      "跳过重试次数（主动放弃）",
		Labels:    []string{"task_type"},
	})

	// MetricServerPanicTotal panic 次数
	MetricServerPanicTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: metricsNamespace,
		Subsystem: serverSubsystem,
		Name:      "panic_total",
		Help:      "panic 次数",
		Labels:    []string{"task_type"},
	})
)

var (
	// MetricSchedulerTriggerTotal 定时任务触发次数
	MetricSchedulerTriggerTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: metricsNamespace,
		Subsystem: serverSubsystem,
		Name:      "scheduler_trigger_total",
		Help:      "定时任务触发次数",
		Labels:    []string{"task_type"},
	})

	// MetricSchedulerRegistered 当前注册的定时任务数
	MetricSchedulerRegistered = metric.NewGaugeVec(&metric.GaugeVecOpts{
		Namespace: metricsNamespace,
		Subsystem: serverSubsystem,
		Name:      "scheduler_registered",
		Help:      "当前注册的定时任务数",
		Labels:    []string{},
	})
)

// ==================== Client 端指标 ====================

const clientSubsystem = "client"

var (
	// MetricClientPushTotal 投递计数
	MetricClientPushTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: metricsNamespace,
		Subsystem: clientSubsystem,
		Name:      "push_total",
		Help:      "投递计数",
		Labels:    []string{"task_type", "push_type", "status"},
	})

	// MetricClientPushDuration 投递耗时
	MetricClientPushDuration = metric.NewHistogramVec(&metric.HistogramVecOpts{
		Namespace: metricsNamespace,
		Subsystem: clientSubsystem,
		Name:      "push_duration_ms",
		Help:      "投递耗时统计(ms)",
		Labels:    []string{"task_type"},
		Buckets:   []float64{1, 2, 5, 10, 25, 50, 100, 250, 500},
	})

	// MetricClientPushBytes 投递字节数
	MetricClientPushBytes = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: metricsNamespace,
		Subsystem: clientSubsystem,
		Name:      "push_bytes",
		Help:      "投递字节数",
		Labels:    []string{"task_type"},
	})

	// MetricClientCancelTotal 撤销任务计数
	MetricClientCancelTotal = metric.NewCounterVec(&metric.CounterVecOpts{
		Namespace: metricsNamespace,
		Subsystem: clientSubsystem,
		Name:      "cancel_total",
		Help:      "撤销任务计数",
		Labels:    []string{"task_type", "status"},
	})
)

// ==================== 队列状态指标（Collector 采集，零侵入） ====================

// QueueMetricsCollector 带队列过滤的指标采集器
// 只采集 allowedQueues 中配置的队列指标，解决多服务共用 Redis 时指标混杂问题
type QueueMetricsCollector struct {
	inspector     *asynq.Inspector
	allowedQueues map[string]struct{} // 允许采集的队列白名单
}

// Descriptors - 与官方 collector 保持一致的指标定义
var (
	tasksQueuedDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricsNamespace, serverSubsystem, "tasks_enqueued_total"),
		"各状态任务数量",
		[]string{"queue", "state"}, nil,
	)

	queueSizeDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricsNamespace, serverSubsystem, "queue_size"),
		"队列任务总数",
		[]string{"queue"}, nil,
	)

	queueLatencyDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricsNamespace, serverSubsystem, "queue_latency_seconds"),
		"队列延迟（最旧 pending 任务等待时间）",
		[]string{"queue"}, nil,
	)

	queueMemUsgDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricsNamespace, serverSubsystem, "queue_memory_usage_approx_bytes"),
		"队列内存占用（采样估算值）",
		[]string{"queue"}, nil,
	)

	tasksProcessedTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricsNamespace, serverSubsystem, "tasks_processed_total"),
		"已处理任务总数（含成功和失败）",
		[]string{"queue"}, nil,
	)

	tasksFailedTotalDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricsNamespace, serverSubsystem, "tasks_failed_total"),
		"失败任务总数",
		[]string{"queue"}, nil,
	)

	pausedQueuesDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricsNamespace, serverSubsystem, "queue_paused_total"),
		"队列暂停状态",
		[]string{"queue"}, nil,
	)

	queueGroupsDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricsNamespace, serverSubsystem, "queue_groups"),
		"聚合组数量",
		[]string{"queue"}, nil,
	)

	tasksAggregatingDesc = prometheus.NewDesc(
		prometheus.BuildFQName(metricsNamespace, serverSubsystem, "tasks_aggregating_total"),
		"聚合中的任务数",
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
		log.Printf("[CRON_METRICS] Failed to collect metrics: %v", err)
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

		// 聚合相关指标
		ch <- prometheus.MustNewConstMetric(queueGroupsDesc, prometheus.GaugeValue, float64(info.Groups), info.Queue)
		ch <- prometheus.MustNewConstMetric(tasksAggregatingDesc, prometheus.GaugeValue, float64(info.Aggregating), info.Queue)
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
			log.Printf("[CRON_METRICS] Failed to get queue info for %s: %v", queueName, err)
			continue
		}
		infos = append(infos, qinfo)
	}
	return infos, nil
}
