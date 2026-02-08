package cron

import (
	"context"
	"crypto/tls"
	"errors"

	"github.com/hibiken/asynq"
)

const (
	ModeSingle   = "single"
	ModeSentinel = "sentinel"
	ModeCluster  = "cluster"
)

type (
	Task struct {
		Type    string
		Payload []byte
	}
	HandlerFunc func(ctx context.Context, t *Task) error

	RedisConf struct {
		//// 模式：single, sentinel, cluster
		Mode             string   `json:",default=single,options=[single,sentinel,cluster]"`
		Username         string   `json:",optional"`
		Password         string   `json:",optional"`
		DialTimeout      int64    `json:",default=5"`                      // 连接建立超时,单位秒，默认5s
		ReadTimeout      int64    `json:",default=3"`                      // 读超时，单位秒，默认3s，-1永不超时，0为默认值
		WriteTimeout     int64    `json:",default=3"`                      // 写超时，单位秒，默认3s，-1永不超时，0为默认值
		DB               int64    `json:",default=0"`                      // Redis DB 索引,默认DB0, mode=cluster无效
		PoolSize         int64    `json:",optional"`                       // 连接池大小，默认值为 10 * cpu 核数
		Network          string   `json:",default=tcp,options=[tcp,unix]"` // tcp 或 unix socket，mode=single有效
		Addr             string   `json:",optional"`                       // Redis 地址 "host:port"，mode=single有效
		MasterName       string   `json:",optional"`                       // 哨兵监控的主节点名，mode=sentinel有效
		SentinelAddrs    []string `json:",optional"`                       // 哨兵节点列表，mode=sentinel有效
		SentinelUsername string   `json:",optional"`                       // 哨兵认证用户名，mode=sentinel有效
		SentinelPassword string   `json:",optional"`                       // 哨兵认证密码，mode=sentinel有效
		Addrs            []string `json:",optional"`                       // 集群种子节点列表，mode=cluster有效
		MaxRedirects     int64    `json:",default=8"`                      // 最大重定向次数，mode=cluster有效
	}
	ServerConfig struct {
		RedisConf
		Namespace string `json:",optional"` // 任务命名空间
		// 任务处理的最大并发量。
		// 注意： 如果设为 0 或负数，NewServer 将会把该值改写为当前进程可用的 CPU 核心数。
		Concurrency int64 `json:",default=0"` // 最大并发任务处理数，默认值为 0（表示使用当前进程可用的 CPU 核心数）
		// 指定当所有队列都为空时，检查新任务的时间间隔。
		// 如果未设置、设为 0 或负数，间隔将被设为 1 秒。
		// 注意： 将此值设置得过低可能会给 Redis 增加显著负载。默认情况下，此值为 1 秒。
		TaskCheckInterval int64 `json:",default=1"` // 检查新任务的时间间隔，默认值为 1 秒,，单位秒
		// 需要处理的任务队列列表及其对应的优先级权重。Key 是队列名称，Value 是关联的优先级数值。
		// 如果设为 nil 或未指定，服务器将只处理 "default" 队列。
		// 优先级处理逻辑（避免低优先级队列饥饿）：
		// Example:
		//
		//     Queues: map[string]int{
		//         "critical": 6,
		//         "default":  3,
		//         "low":      1,
		//     }
		//
		// 在所有队列都不为空的情况下，"critical"、"default"、"low" 中的任务被处理的时间比例应分别为 60%、30% 和 10%。
		// 如果某个队列的优先级值为 0 或负数，该队列将被忽略。
		Queues map[string]int `json:",optional"` // 任务队列列表及其对应的优先级权重，如果为nil,则只处理"default"队列, 默认值 map["default":1]
		// 指示是否严格对待队列优先级。
		// 若为 true，只有当高优先级队列为空时才处理低优先级。
		StrictPriority bool `json:",default=false"` // 是否严格对待队列优先级，默认值为 false
		// 在强行中止 Worker 之前，等待其完成任务的时长。
		// 若超时，任务会重新推入 Redis。
		// 如果设为 0 间隔将被设为 8 秒。
		ShutdownTimeout int64 `json:",default=8"` // 强行中止 Worker 之前，等待其完成任务的时长，默认值为 8 秒，单位秒
		// 健康检查（Ping）的时间间隔
		// 如果设为 0 ，间隔将被设为 15 秒。
		// 可以触发 HealthCheckFunc
		HealthCheckInterval int64 `json:",default=15"` // 健康检查（Ping）的时间间隔，默认值为 15 秒，单位秒
		// 检查「已计划」和「重试」任务并将其转为「待处理」状态的时间间隔。
		// 如果设为 0 ，间隔将被设为 5 秒。
		DelayedTaskCheckInterval int64 `json:",default=5"` // 检查「已计划」和「重试」任务并将其转为「待处理」状态的时间间隔，默认值为 5 秒，单位秒
		// 服务器在聚合组内任务前等待新任务进入的时间。最小值为 1 秒，若小于 1 秒会引发 Panic。
		// 如果设置为0,则默认使用1分钟,即60s
		GroupGracePeriod int64 `json:",default=60"` // 任务分组 grace period，默认值为 60 秒，单位秒
		// 在聚合组内任务前，服务器等待新任务进入的最长时间。
		// 如果未设置、则无限制
		GroupMaxDelay int64 `json:",default=0"` // 任务分组最大延迟时间，默认值为 0 秒，单位秒
		//	一个组内可聚合为单个任务的最大任务数量。达到此值会立即触发聚合
		// 如果设为 0 ，则无限制
		GroupMaxSize int64 `json:",default=0"` // 一个组内可聚合为单个任务的最大任务数量。达到此值会立即触发聚合，默认值为 0，表示无限制
		// 任务清理器（Janitor）检查过期任务的时间间隔。
		// 如果未设置、设为 0 或负数，间隔将被设为 8 秒。
		JanitorInterval int64 `json:",default=8"` // 任务清理器（Janitor）检查过期任务的时间间隔，默认值为 8 秒，单位秒
		// 单次清理运行中删除的任务数量。建议不要设置太大以防脚本长时间运行。
		// 如果为0,则使用默认值100
		JanitorBatchSize int64 `json:",default=100"`
	}

	ClientConfig struct {
		RedisConf
	}
)

func (r *RedisConf) buildRedisOpts(tlsConfig *tls.Config) (asynq.RedisConnOpt, error) {
	mode := r.Mode
	switch mode {
	case ModeSingle:
		return asynq.RedisClientOpt{
			Network:      r.Network,
			Addr:         r.Addr,
			Username:     r.Username,
			Password:     r.Password,
			DB:           int(r.DB),
			DialTimeout:  toDuration(r.DialTimeout),
			ReadTimeout:  toDuration(r.ReadTimeout),
			WriteTimeout: toDuration(r.WriteTimeout),
			PoolSize:     int(r.PoolSize),
			TLSConfig:    tlsConfig,
		}, nil
	case ModeCluster:
		return asynq.RedisClusterClientOpt{
			Addrs:        r.Addrs,
			MaxRedirects: int(r.MaxRedirects),
			Username:     r.Username,
			Password:     r.Password,
			DialTimeout:  toDuration(r.DialTimeout),
			ReadTimeout:  toDuration(r.ReadTimeout),
			WriteTimeout: toDuration(r.WriteTimeout),
			TLSConfig:    tlsConfig,
		}, nil
	case ModeSentinel:
		return asynq.RedisFailoverClientOpt{
			MasterName:       r.MasterName,
			SentinelAddrs:    r.SentinelAddrs,
			SentinelUsername: r.SentinelUsername,
			SentinelPassword: r.SentinelPassword,
			Username:         r.Username,
			Password:         r.Password,
			DB:               int(r.DB),
			DialTimeout:      toDuration(r.DialTimeout),
			ReadTimeout:      toDuration(r.ReadTimeout),
			WriteTimeout:     toDuration(r.WriteTimeout),
			PoolSize:         int(r.PoolSize),
			TLSConfig:        tlsConfig,
		}, nil
	default:
		return nil, errors.New("invalid mode")
	}
}

func (r *RedisConf) Validate() error {
	switch r.Mode {
	case ModeSingle:
		if r.Addr == "" {
			return errors.New("single mode addr cannot be empty")
		}
	case ModeSentinel:
		if r.MasterName == "" {
			return errors.New("sentinel mode master name cannot be empty")

		}
		if len(r.SentinelAddrs) == 0 {
			return errors.New("sentinel mode sentinel addrs cannot be empty")
		}
	case ModeCluster:
		if len(r.Addrs) == 0 {
			return errors.New("cluster mode addrs cannot be empty")
		}
	default:
		return errors.New("invalid redis mode")
	}

	return nil
}

func (s *ServerConfig) Validate() error {
	if err := s.RedisConf.Validate(); err != nil {
		return err
	}
	if s.Concurrency < 0 {
		return errors.New("concurrency must be a positive integer")
	}

	if s.GroupGracePeriod < 0 {
		return errors.New("group grace period must be a positive integer")
	}

	if s.GroupGracePeriod > 0 && s.GroupGracePeriod < 1 {
		return errors.New("group grace period must be greater than or equal to 1 second")
	}

	if s.GroupMaxSize < 0 {
		return errors.New("group max size must be a positive integer")
	}

	if s.GroupMaxDelay < 0 {
		return errors.New("group max delay must be a positive integer")
	}

	if s.JanitorInterval < 0 {
		return errors.New("janitor interval must be a positive integer")
	}

	if s.JanitorBatchSize < 0 {
		return errors.New("janitor batch size must be a positive integer")
	}

	if s.DelayedTaskCheckInterval < 0 {
		return errors.New("delayed task check interval must be a positive integer")
	}

	return nil
}
