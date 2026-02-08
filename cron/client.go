package cron

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

type (
	ClientOption func(c *CommonClient)
	Client       interface {
		Close() error
		Push(ctx context.Context, taskType string, payload []byte, opts ...asynq.Option) (*asynq.TaskInfo, error)
		PushIn(ctx context.Context, taskType string, payload []byte, delay time.Duration, opts ...asynq.Option) (*asynq.TaskInfo, error)
		PushAt(ctx context.Context, taskType string, payload []byte, at time.Time, opts ...asynq.Option) (*asynq.TaskInfo, error)
		PushJson(ctx context.Context, taskType string, data any, opts ...asynq.Option) (*asynq.TaskInfo, error)
		PushInJson(ctx context.Context, taskType string, data any, delay time.Duration, opts ...asynq.Option) (*asynq.TaskInfo, error)
		PushAtJson(ctx context.Context, taskType string, data any, at time.Time, opts ...asynq.Option) (*asynq.TaskInfo, error)
		CancelTask(queue, taskID string) error
		RescheduleTask(ctx context.Context, queue, taskID string, taskType string, data any, newDelay time.Duration, opts ...asynq.Option) (*asynq.TaskInfo, error)
	}

	CommonClient struct {
		conf      ClientConfig
		client    *asynq.Client
		inspector *asynq.Inspector
		tlsConfig *tls.Config
	}
)

func MustNewClient(conf ClientConfig, opts ...ClientOption) *CommonClient {
	c, err := NewClient(conf, opts...)
	logx.Must(err)
	return c
}

func NewClient(conf ClientConfig, opts ...ClientOption) (*CommonClient, error) {
	if err := conf.RedisConf.Validate(); err != nil {
		return nil, err
	}
	c := &CommonClient{
		conf: conf,
	}
	for _, opt := range opts {
		opt(c)
	}

	redisClientOpts, err := conf.RedisConf.buildRedisOpts(c.tlsConfig)
	if err != nil {
		return nil, err
	}
	c.client = asynq.NewClient(redisClientOpts)
	c.inspector = asynq.NewInspector(redisClientOpts)
	return c, nil
}

func MustNewClientFromRedisClient(rds redis.UniversalClient) *CommonClient {
	c, err := NewClientFromRedisClient(rds)
	logx.Must(err)
	return c
}

func NewClientFromRedisClient(rds redis.UniversalClient) (*CommonClient, error) {
	c := &CommonClient{
		client:    asynq.NewClientFromRedisClient(rds),
		inspector: asynq.NewInspectorFromRedisClient(rds),
	}
	return c, nil
}

// buildTask 处理 Trace 注入
func (c *CommonClient) buildTask(ctx context.Context, taskType string, payload []byte) *asynq.Task {
	header := make(map[string]string)
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(header))
	return asynq.NewTaskWithHeaders(taskType, payload, header)
}

// push 核心推送逻辑，处理埋点和入队
func (c *CommonClient) push(ctx context.Context, taskType string, payload []byte, opts ...asynq.Option) (info *asynq.TaskInfo, err error) {
	ctx, span := StartProducerSpan(ctx, taskType)
	defer func() {
		EndSpan(span, err)
	}()

	task := c.buildTask(ctx, taskType, payload)
	info, err = c.client.EnqueueContext(ctx, task, opts...)
	if err == nil && info != nil {
		span.SetAttributes(attribute.String("messaging.message_id", info.ID))
	}
	return info, err
}

// Push 立即执行
func (c *CommonClient) Push(ctx context.Context, taskType string, payload []byte, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.push(ctx, taskType, payload, opts...)
}

// PushIn 延时执行
func (c *CommonClient) PushIn(ctx context.Context, taskType string, payload []byte, delay time.Duration, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.push(ctx, taskType, payload, append(opts, asynq.ProcessIn(delay))...)
}

// PushAt 指定时间点执行
func (c *CommonClient) PushAt(ctx context.Context, taskType string, payload []byte, at time.Time, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.push(ctx, taskType, payload, append(opts, asynq.ProcessAt(at))...)
}

// PushJson 立即执行
func (c *CommonClient) PushJson(ctx context.Context, taskType string, data any, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("asynq marshal error: %w", err)
	}
	return c.push(ctx, taskType, payload, opts...)
}

// PushInJson 延时执行
func (c *CommonClient) PushInJson(ctx context.Context, taskType string, data any, delay time.Duration, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("asynq marshal error: %w", err)
	}
	return c.push(ctx, taskType, payload, append(opts, asynq.ProcessIn(delay))...)
}

// PushAtJson 指定时间点执行
func (c *CommonClient) PushAtJson(ctx context.Context, taskType string, data any, at time.Time, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("asynq marshal error: %w", err)
	}
	return c.push(ctx, taskType, payload, append(opts, asynq.ProcessAt(at))...)
}

// CancelTask 撤回任务
// 只能撤回处于 Scheduled(延时), Pending(排队), Retry(重试) 状态的任务
func (c *CommonClient) CancelTask(queue, taskID string) error {
	// 如果不指定 queue，asynq 默认是 "default"
	if queue == "" {
		queue = "default"
	}

	// DeleteTask 会从所有等待队列中尝试删除该 ID
	err := c.inspector.DeleteTask(queue, taskID)
	if err != nil {
		// 如果任务已经开始执行(Active)或已完成，删除会报错
		return fmt.Errorf("asynq cancel task [%s] failed: %w", taskID, err)
	}
	return nil
}

// RescheduleTask 调整任务执行时间 (先删后加)
func (c *CommonClient) RescheduleTask(ctx context.Context, queue, taskID string, taskType string, data any, newDelay time.Duration, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	// 1. 尝试撤回旧任务
	// 注意：如果任务不存在或已在运行，我们通常记录日志但继续发送新任务，或者根据业务抛错
	err := c.CancelTask(queue, taskID)
	if err != nil {
		return nil, err
	}

	// 2. 重新发送新任务，并强制带上原有的 TaskID 保证唯一性
	newOpts := append(opts, asynq.TaskID(taskID))
	return c.PushInJson(ctx, taskType, data, newDelay, newOpts...)
}

func (c *CommonClient) Close() error {
	err := c.client.Close()
	_ = c.inspector.Close()
	return err
}

func WithClientTLS(tlsCfg *tls.Config) ClientOption {
	return func(c *CommonClient) {
		c.tlsConfig = tlsCfg
	}
}
