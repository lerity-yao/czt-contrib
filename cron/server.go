package cron

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/hibiken/asynq/x/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/zeromicro/go-zero/core/logx"
)

type (
	ServerOption func(c *CommonServer)

	Server interface {
		Start()
		Stop()
		Add(pattern string, handler HandlerFunc)
		CronAdd(spec string, pattern string, opts ...asynq.Option) string
	}

	CommonServer struct {
		conf            ServerConfig
		redisClientOpts asynq.RedisConnOpt
		config          asynq.Config
		Server          *asynq.Server
		Mux             *asynq.ServeMux
		Scheduler       *asynq.Scheduler
		inspector       *asynq.Inspector
		tlsConfig       *tls.Config
	}
)

func MustNewServer(conf ServerConfig, opts ...ServerOption) Server {
	s, err := NewServer(conf, opts...)
	logx.Must(err)
	return s
}

func NewServer(conf ServerConfig, opts ...ServerOption) (Server, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}

	s := &CommonServer{conf: conf}
	s.buildConfig()
	for _, opt := range opts {
		opt(s)
	}
	redisClientOpts, err := s.conf.RedisConf.buildRedisOpts(s.tlsConfig)
	if err != nil {
		return nil, err
	}
	s.redisClientOpts = redisClientOpts
	s.buildPrometheus()
	s.Scheduler = asynq.NewScheduler(s.redisClientOpts, &asynq.SchedulerOpts{
		Location: time.Local,
		Logger:   s.config.Logger,
	})
	s.Server = asynq.NewServer(s.redisClientOpts, s.config)
	s.Mux = asynq.NewServeMux()
	s.Mux.Use(RecoveryMiddleware)
	s.Mux.Use(PrometheusMiddleware)
	s.Mux.Use(TraceMiddleware)
	s.Mux.Use(LoggingMiddleware)
	return s, nil
}

func (c *CommonServer) buildConfig() {
	queues := c.conf.Queues
	if len(queues) == 0 && c.conf.Namespace != "" {
		queues = map[string]int{
			c.conf.Namespace: 1, // 默认优先级为 1
		}
	}
	c.config = asynq.Config{
		Concurrency:              int(c.conf.Concurrency),
		TaskCheckInterval:        toDuration(c.conf.TaskCheckInterval),
		Queues:                   queues,
		StrictPriority:           c.conf.StrictPriority,
		ShutdownTimeout:          toDuration(c.conf.ShutdownTimeout),
		HealthCheckInterval:      toDuration(c.conf.HealthCheckInterval),
		DelayedTaskCheckInterval: toDuration(c.conf.DelayedTaskCheckInterval),
		GroupGracePeriod:         toDuration(c.conf.GroupGracePeriod),
		GroupMaxDelay:            toDuration(c.conf.GroupMaxDelay),
		GroupMaxSize:             int(c.conf.GroupMaxSize),
		JanitorInterval:          toDuration(c.conf.JanitorInterval),
		JanitorBatchSize:         int(c.conf.JanitorBatchSize),
	}
}

func (c *CommonServer) buildPrometheus() {
	c.inspector = asynq.NewInspector(c.redisClientOpts)
	// 2. 初始化官方采集器
	collector := metrics.NewQueueMetricsCollector(c.inspector)
	_ = prometheus.Register(collector)
}

func (c *CommonServer) ServerMux() *asynq.ServeMux {
	return c.Mux
}

// Add 添加任务处理函数
func (c *CommonServer) Add(pattern string, handler HandlerFunc) {
	realPattern := pattern
	if c.conf.Namespace != "" {
		realPattern = fmt.Sprintf("%s:%s", c.conf.Namespace, pattern)
	}
	asynqHandler := func(ctx context.Context, at *asynq.Task) error {
		p := at.Payload()
		pc := make([]byte, len(p))
		copy(pc, p)

		bt := &Task{
			Type:    at.Type(),
			Payload: pc, // 使用拷贝后的数据
		}
		return handler(ctx, bt)
	}

	// 注册到原生的 Mux
	c.Mux.HandleFunc(realPattern, asynqHandler)
	logx.Infof("[ASYNQ] 注册外部任务(待命): %s", realPattern)

}

// CronAdd 注册定时任务 (Server 自产自销)
func (c *CommonServer) CronAdd(spec string, pattern string, opts ...asynq.Option) string {
	realPattern := pattern
	finalOpts := opts
	if c.conf.Namespace != "" {
		realPattern = fmt.Sprintf("%s:%s", c.conf.Namespace, pattern)
		finalOpts = append(opts, asynq.Queue(c.conf.Namespace))
	}
	task := asynq.NewTask(realPattern, nil, finalOpts...)
	entryID, err := c.Scheduler.Register(spec, task, finalOpts...)
	if err != nil {
		logx.Errorf("[ASYNQ] 自动注册定时任务失败: type=%s, spec=%s, err=%v", realPattern, spec, err)
	}

	logx.Infof("[ASYNQ] Cron job registered: [%s] -> %s (EntryID: %s)", spec, realPattern, entryID)
	return entryID
}

// Start 启动任务处理服务
func (c *CommonServer) Start() {
	if err := c.Scheduler.Start(); err != nil {
		logx.Must(err)
	}
	if err := c.Server.Run(c.Mux); err != nil {
		logx.Must(err)
	}
}

// Stop 停止任务处理服务
func (c *CommonServer) Stop() {
	c.Scheduler.Shutdown()
	c.Server.Shutdown()
	if c.inspector != nil {
		_ = c.inspector.Close()
	}
}

func WithServerTLS(tlsCfg *tls.Config) ServerOption {
	return func(c *CommonServer) {
		c.tlsConfig = tlsCfg
	}
}

func WithServerLogger(logger asynq.Logger) ServerOption {
	return func(c *CommonServer) {
		c.config.Logger = logger
	}
}
