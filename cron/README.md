# cron

[中文](./readme-cn.md)

A distributed task queue system built on [Asynq](https://github.com/hibiken/asynq), designed as a scheduled task and asynchronous task processing module for the Go-Zero framework.

## Features

- 🚀 **High Performance**: High-performance distributed task queue based on Redis
- ⏰ **Scheduled Tasks**: Supports Cron expression scheduled tasks
- 🔄 **Asynchronous Processing**: Asynchronous task queue with delayed execution support
- 📊 **Metrics**: Built-in Prometheus metrics collection
- 🔍 **Distributed Tracing**: Integrated OpenTelemetry tracing
- 🛡️ **Error Recovery**: Automatic panic recovery and error handling
- 🔧 **Flexible Configuration**: Supports multiple Redis modes (Single, Sentinel, Cluster)

## Installation

```bash
go get github.com/lerity-yao/czt-contrib/cron
```
## Service Generation

Supports one-click server code generation similar to the goctl tool. The tool is `cztctl`, a customized version of goctl.

You can also customize the server code generation templates.

Please refer to [cztctl](https://github.com/lerity-yao/czt-contrib/blob/main/cztctl/README.md).

## ⚙️ Configuration Parameters

### RedisConf (Basic Connection Configuration)
This configuration controls how to connect to Redis, supporting Single, Sentinel, and Cluster modes.

| Parameter | Type | Default | Description and Recommendations |
| --- | --- | --- | --- |
| Mode | string | single | Required. Options: single, sentinel, cluster. Determines which subsequent fields take effect. |
| Addr | string | - | Required when Mode=single. Format "host:port". |
| Addrs | []string | - | Required when Mode=cluster. List of cluster seed nodes; only some nodes are needed, and the driver will auto-discover the full topology. |
| MasterName | string | - | Required when Mode=sentinel. The master node name monitored in Sentinel mode (usually defaults to mymaster). |
| SentinelAddrs | []string | - | Required when Mode=sentinel. List of Sentinel nodes. At least 3 are recommended for high availability. |
| Username | string | - | Redis 6.0+ ACL authentication username. |
| Password | string | - | Redis authentication password. |
| DB | int64 | 0 | Redis database index. Note: this is invalid in Cluster mode. |
| PoolSize | int64 | - | Maximum number of connections in the pool. Defaults to 10 * number of CPU cores. For high-concurrency tasks, increase based on concurrency. |
| DialTimeout | int64 | 5 | Connection establishment timeout in seconds. Increase appropriately in poor network environments. |
| ReadTimeout | int64 | 3 | Read timeout in seconds. Recommended to keep the default. |
| WriteTimeout | int64 | 3 | Write timeout in seconds. Recommended to keep the default. |

### ServerConfig (Task Processing Engine Configuration)

This configuration directly affects consumer processing efficiency, stability, and resource usage.

- Namespace (string):
Core logic: all keys in Redis will be prefixed with this namespace.
Recommendation: use a different Namespace for each independent service. This provides physical isolation and prevents workers of different services from accidentally consuming each other's tasks.


- Concurrency (int64):
Default: 0 (automatically set to the number of CPU cores).
Recommendation: if tasks involve a lot of network IO (e.g., sending SMS, calling third-party APIs), increase to 20~100; for CPU-intensive tasks, keep the default or increase slightly.


- Queues (map[string]int):
Core logic: defines which queues to listen to and their weights.
Practical example: {"critical": 6, "default": 3, "low": 1} means 60% of effort goes to critical tasks.

- StrictPriority (bool):
Logic: if true, as long as the critical queue has a task, the worker will never touch the default queue.
Note: enabling this may cause low-priority queues to starve (never get processed); use with caution.

- TaskCheckInterval (int64):
Logic: how long the worker rests before checking Redis again when all queues are empty.
Recommendation: default 1 second. Too short increases Redis CPU load; too long causes noticeable task processing delays.


- ShutdownTimeout (int64):
Logic: maximum time the worker waits for current tasks to complete during graceful shutdown.
Recommendation: default 8 seconds. If your task logic is long (e.g., processing large files), you must increase this value, otherwise tasks will be forcibly interrupted and requeued.


- DelayedTaskCheckInterval (int64):
Logic: frequency of checking whether delayed tasks and retry tasks are due. Default 5 seconds.


- HealthCheckInterval (int64):
Logic: heartbeat detection between the worker and Redis. Recommended to keep the default 15 seconds.


- GroupGracePeriod (int64): Aggregation window (sliding window). Default 60 seconds. The timer resets each time a new task enters the group; aggregation is triggered only when no new tasks arrive during the grace period.


- GroupMaxDelay (int64): Maximum waiting time to force aggregation (hard limit). Default 300 seconds. From the arrival of the first task in the group, aggregation is forced at this time regardless of new tasks. Prevents GracePeriod from being reset indefinitely and never aggregating.


- GroupMaxSize (int64): When the number of tasks in the group reaches this threshold, aggregation is triggered immediately without waiting for the window. Default 0 (unlimited).


> The above three conditions are in an **OR** relationship; aggregation is triggered when any one is met. Must be used with `WithGroupAggregator`; otherwise the grouping feature does not take effect.


- JanitorInterval (int64): Interval for checking and cleaning completed/expired tasks in Redis.


- JanitorBatchSize (int64): Maximum number of items deleted per cleanup operation. Default 100. Prevents Redis from blocking due to deleting too many at once.

**Configuration Recommendations**
- Namespace must be set: this is the foundation for multiple services to coexist.
- Set Concurrency reasonably: larger for more IO, smaller for more CPU.
- Set ShutdownTimeout: must be greater than the longest possible execution time in your business logic.

### ClientConfig (Client Configuration)

Contains only `RedisConf`, sharing the same Redis connection configuration as the Server.

### ServerOption

| Option | Parameter | Description |
|--------|------|------|
| `WithServerTLS` | `*tls.Config` | Set Redis TLS connection configuration |
| `WithServerLogger` | `asynq.Logger` | Replace the default logger (recommended: `&cron.AsynqLogger{}` to bridge to go-zero logx) |
| `WithGroupAggregator` | `asynq.GroupAggregator` | Inject a group aggregator to enable Group functionality |
| `WithRetryDelayFunc` | `asynq.RetryDelayFunc` | Customize retry backoff strategy (default: `ExponentialRetryDelay`) |

### ClientOption

| Option | Parameter | Description |
|--------|------|------|
| `WithClientTLS` | `*tls.Config` | Set Redis TLS connection configuration |


## API Reference

### Constructors

| Function | Signature | Description |
|------|------|------|
| `MustNewServer` | `func MustNewServer(conf ServerConfig, opts ...ServerOption) Server` | Create a Server; panics on failure |
| `NewServer` | `func NewServer(conf ServerConfig, opts ...ServerOption) (Server, error)` | Create a Server; returns error on failure |
| `MustNewClient` | `func MustNewClient(conf ClientConfig, opts ...ClientOption) *CommonClient` | Create a Client; panics on failure |
| `NewClient` | `func NewClient(conf ClientConfig, opts ...ClientOption) (*CommonClient, error)` | Create a Client; returns error on failure |
| `MustNewClientFromRedisClient` | `func MustNewClientFromRedisClient(rds redis.UniversalClient) *CommonClient` | Create a Client from an existing Redis connection; panics on failure |
| `NewClientFromRedisClient` | `func NewClientFromRedisClient(rds redis.UniversalClient) (*CommonClient, error)` | Create a Client from an existing Redis connection; returns error on failure |

### Public Types

| Type | Definition | Description |
|------|------|------|
| `Task` | `struct { Type string; Payload []byte }` | Task carrier received by the handler; Type is the task type, Payload is the raw byte data |
| `HandlerFunc` | `func(ctx context.Context, t *Task) error` | Task handler function signature; all handlers must implement this type |
| `AsynqLogger` | `struct{}` | Built-in log adapter that bridges asynq logs to go-zero logx |

### Server Interface Methods

| Method | Signature | Description |
|------|------|------|
| `Add` | `Add(pattern string, handler HandlerFunc)` | Register a task handler (consume tasks pushed externally); automatically prepends the Namespace prefix |
| `CronAdd` | `CronAdd(spec string, pattern string, handler HandlerFunc, opts ...asynq.Option) string` | One-step scheduled task registration (self-produced and self-consumed); completes both handler registration and scheduled dispatch, automatically sets TaskID deduplication, returns EntryID |
| `SetBaseContext` | `SetBaseContext(ctx context.Context)` | Inject a base context; all task handler ctxs use this as parent. Must be called before Start() |
| `Start` | `Start()` | Start Scheduler + Processor; non-blocking |
| `Stop` | `Stop()` | Graceful shutdown: closes Scheduler → Server → Inspector in order |

### Client Interface Methods

| Method | Signature | Description |
|------|------|------|
| `Push` | `Push(ctx, taskType string, payload []byte, opts ...asynq.Option) (*asynq.TaskInfo, error)` | Execute immediately; push raw bytes |
| `PushJson` | `PushJson(ctx, taskType string, data any, opts ...asynq.Option) (*asynq.TaskInfo, error)` | Execute immediately; auto JSON serialization |
| `PushIn` | `PushIn(ctx, taskType string, payload []byte, delay time.Duration, opts ...) (*asynq.TaskInfo, error)` | Delayed execution; specify Duration |
| `PushInJson` | `PushInJson(ctx, taskType string, data any, delay time.Duration, opts ...) (*asynq.TaskInfo, error)` | Delayed execution; auto JSON serialization |
| `PushAt` | `PushAt(ctx, taskType string, payload []byte, at time.Time, opts ...) (*asynq.TaskInfo, error)` | Execute at a specific time; specify absolute time |
| `PushAtJson` | `PushAtJson(ctx, taskType string, data any, at time.Time, opts ...) (*asynq.TaskInfo, error)` | Execute at a specific time; auto JSON serialization |
| `CancelTask` | `CancelTask(queue, taskID string) error` | Withdraw tasks in Scheduled/Pending/Retry state |
| `RescheduleTask` | `RescheduleTask(ctx, queue, taskID, taskType string, data any, newDelay time.Duration, opts ...) (*asynq.TaskInfo, error)` | Atomic withdraw + re-dispatch; TaskID remains unchanged |
| `Close` | `Close() error` | Close the client connection |

> All Push methods default to `MaxRetry=0` (no retries); this can be overridden via opts.

### Utility Functions

| Function | Signature | Description |
|------|------|------|
| `ExponentialRetryDelay` | `func(n int, _ error, _ *asynq.Task) time.Duration` | Built-in exponential backoff (`2^n - 1` seconds); used by Server by default |
| `GetTaskID` | `func(ctx context.Context) (string, bool)` | Task ID; remains unchanged across retries |
| `GetRetryCount` | `func(ctx context.Context) (int, bool)` | Current retry count; 0 = first execution |
| `GetMaxRetry` | `func(ctx context.Context) (int, bool)` | Maximum retry limit |
| `GetQueueName` | `func(ctx context.Context) (string, bool)` | Queue name where the task resides |

## Advanced Guide

### Namespace and TaskType Matching Rules

**This is an important note: when pushing tasks from the client, you must understand the server's Namespace configuration.**

Configuring `Namespace` on the server has two effects:

1. **taskType automatic concatenation**: the `Add` method converts the pattern to the `Namespace:pattern` format
2. **Queue automatic creation**: the server listens on queues named after the Namespace

```go
// server.go Add 方法核心逻辑
func (c *CommonServer) Add(pattern string, handler HandlerFunc) {
    realPattern := pattern
    if c.conf.Namespace != "" {
        realPattern = fmt.Sprintf("%s:%s", c.conf.Namespace, pattern)
    }
    c.Mux.HandleFunc(realPattern, asynqHandler)
}

// server.go buildConfig 方法核心逻辑
func (c *CommonServer) buildConfig() {
    queues := c.conf.Queues
    if len(queues) == 0 && c.conf.Namespace != "" {
        queues = map[string]int{
            c.conf.Namespace: 1,  // 服务端监听以 Namespace 命名的队列
        }
    }
    // ...
}
```

**In the go-zero environment**, the service name is usually used as the Namespace:

```go
// serviceContext.go 中常见写法
c.WorkConf.Namespace = c.Name  // 使用 go-zero 配置的服务名
```

**When pushing tasks from the client, two conditions must be met simultaneously:**

1. **taskType carries the Namespace prefix**: consistent with the full pattern registered on the server
2. **Specify the queue**: use `asynq.Queue(namespace)` to specify which queue the task is pushed to

```go
import "github.com/hibiken/asynq"

// Assume server config: Namespace = "order-service", registered: Add("send_email", handler)
// Actual registered pattern on the server: "order-service:send_email"
// Queue listened by the server: "order-service"

// ✔️ Correct: taskType with namespace + queue specified
client.PushJson(ctx, "order-service:send_email", payload, asynq.Queue("order-service"))
```

**Summary**:
- Unify Namespace naming conventions between client and server
- Define constants to manage taskType and Namespace centrally and avoid hard-coding
- If the server does not configure Namespace, use the original pattern directly as taskType and there is no need to specify a queue (defaults to "default")

### Retry Backoff Strategy

The Server uses `ExponentialRetryDelay` exponential backoff by default; the delay formula is `2^n - 1` seconds:

| Retry Count | Delay |
|---------|------|
| 1 | 1s |
| 2 | 3s |
| 3 | 7s |
| 4 | 15s |
| 5 | 31s |
| 6 | 63s (~1 minute) |
| 7 | 127s (~2 minutes) |
| 10 | 1023s (~17 minutes) |

To customize the backoff strategy, override it via `WithRetryDelayFunc`:

```go
server := cron.MustNewServer(c.WorkConf,
    cron.WithRetryDelayFunc(func(n int, e error, t *asynq.Task) time.Duration {
        // 固定间隔 10 秒
        return 10 * time.Second
    }),
)
```

#### Task Metadata

In the handler, task runtime metadata can be obtained from `ctx`:

```go
retryCount, _ := cron.GetRetryCount(ctx)  // 当前第几次重试（0 = 首次执行）
maxRetry, _   := cron.GetMaxRetry(ctx)    // 最大重试次数
taskID, _     := cron.GetTaskID(ctx)      // 任务 ID（重试不变）
queueName, _  := cron.GetQueueName(ctx)   // 队列名
```

#### Skip Retry

When an error is unrecoverable (invalid parameters, business rejection, idempotency conflict), use `asynq.SkipRetry` to skip retries:

```go
// internal/logic/demoA/gdemoalogic.go
import "github.com/hibiken/asynq"

func (l *GDemoALogic) GDemoA(req *types.Name) error {
    if req.Name == "" {
        // 参数非法，重试也没用
        return fmt.Errorf("invalid name: %w", asynq.SkipRetry)
    }
    // ...
    return nil
}
```

> **Decision criterion:** if retrying will yield the same result → SkipRetry; if retrying might succeed → let it retry.
>
> **Note:** must wrap with `fmt.Errorf("%w", asynq.SkipRetry)`. `github.com/pkg/errors`'s `errors.Wrap` is incompatible with the standard library `errors.Is()` unwrap chain, so asynq cannot recognize it.

### Task Timeout Control

> **From v0.1.0**: execution timeouts for scheduled tasks and pushed tasks can be injected at the **per-task granularity** by the business side; the `ctx` received by the handler carries a `Deadline`, which automatically triggers `ctx.Done()` when expired, allowing the business to exit proactively.

#### Design Points

- asynq's default timeout is a hard-coded constant of **30 minutes**, and cannot be configured globally via `asynq.Config`.
- Per-task timeout is written to `msg.Timeout` via the enqueue metadata `asynq.Timeout(d)`; when the processor pulls a task, it injects ctx via `context.WithDeadline(baseCtx, deadline)`.
- Only enqueue actions (`CronAdd`, `Client.Push*`) can carry `asynq.Timeout`; pure handler registration via `Add` does not participate in dispatch and cannot set it.

#### Scheduled Tasks: Managed by go-zero RestConf.Timeout

`rest.RestConf.Timeout` defaults to `3000ms` and must be > 0; it can be reused directly in `workers.go` as the scheduled task timeout baseline:

```go
// internal/handler/workers.go
import (
    "time"
    "github.com/hibiken/asynq"
)

func RegisterHandlers(server *service.ServiceGroup, serverCtx *svc.ServiceContext) {
    var taskOpts []asynq.Option
    taskOpts = append(taskOpts,
        asynq.Timeout(time.Duration(serverCtx.Config.Timeout)*time.Millisecond),
        asynq.MaxRetry(0),
    )

    serverCtx.CronServer.CronAdd("*/1 * * * *", "GDemoA",
        demoA.GDemoAHandler(serverCtx), taskOpts...)

    server.Add(serverCtx.CronServer)
}
```

Effect: when the scheduled time arrives and the task is enqueued, `msg.Timeout` is automatically set to the `Timeout` configured in yaml; the handler receives ctx with `Deadline = now + Timeout`. Changing yaml globally adjusts the timeout limit for all scheduled tasks without code changes.

> The default 3 seconds is short for some IO-intensive scheduled tasks; it is recommended to configure explicitly in yaml (e.g., `Timeout: 30000`).

#### Pushed Tasks: Custom Timeout per Task on the Client

The `Client.Push*` family all support passing through `asynq.Timeout(d)`, allowing the pusher to specify as needed based on actual task duration, **not limited by the 30-minute default**:

```go
import "github.com/hibiken/asynq"

// Short task: 5-second timeout
client.PushJson(ctx, "order-service:send_sms", payload,
    asynq.Queue("order-service"),
    asynq.Timeout(5*time.Second))

// Long task: 1-hour timeout
client.PushJson(ctx, "order-service:export_report", payload,
    asynq.Queue("order-service"),
    asynq.Timeout(time.Hour))
```

> If the pusher does not pass `asynq.Timeout`, it falls back to the asynq default of 30 minutes. In production, it is recommended to set each task explicitly according to its time budget.

### Task Group Aggregation

Merge multiple tasks in the same group into a single batch task before processing; suitable for notification merging, batch writes, and similar scenarios.

#### Trigger Conditions (Any One of Three)

| Condition | Config Parameter | Default | Trigger Mechanism |
|------|----------|--------|----------|
| Grace period timeout | `GroupGracePeriod` | 60s | **Sliding window**: resets the timer each time a new task arrives; aggregation is triggered only when no new tasks arrive during the grace period |
| Maximum delay | `GroupMaxDelay` | 300s | **Hard limit**: from the arrival of the first task, aggregation is forced at this time as a safeguard |
| Count threshold | `GroupMaxSize` | 0 (unlimited) | When the number of tasks in the group reaches the threshold, aggregate immediately |

#### Trigger Scenario Examples

Assume configuration: `GracePeriod=60s, MaxDelay=300s, MaxSize=50`

| Scenario | Behavior |
|------|------|
| 3 tasks arrive and then no more | Aggregated 60s after the last task arrives (hits GracePeriod) |
| Tasks keep arriving every 30s | Forced aggregation at 300s (hits MaxDelay safeguard) |
| 50 tasks arrive in quick succession | Aggregate immediately (hits MaxSize) |
| One task arrives every 61s | Each task aggregates separately (61s > 60s, GracePeriod triggers first) |

#### Usage

**1. Inject the aggregator on the server (required)**

```go
import "github.com/hibiken/asynq"

server := cron.MustNewServer(c.WorkConf,
    cron.WithGroupAggregator(asynq.GroupAggregatorFunc(func(group string, tasks []*asynq.Task) *asynq.Task {
        // 根据 group 名路由不同的聚合逻辑
        switch group {
        case "batch_email":
            return mergeEmails(tasks)
        case "batch_order":
            return mergeOrders(tasks)
        default:
            return defaultMerge(tasks)
        }
    })),
)
```

> The whole Server has only one `GroupAggregator`; different group aggregation logics are distinguished by the `group` parameter.

**2. Specify the group when pushing from the client**

```go
import "github.com/hibiken/asynq"

// Push to the "batch_email" group
client.PushJson(ctx, "order-service:send_email", payload,
    asynq.Queue("order-service"),
    asynq.Group("batch_email"),  // 指定分组名
)
```

**3. Register the handler for the merged task on the server**

```go
// internal/handler/batch/batchSendEmailHandler.go
func BatchSendEmailHandler(svcCtx *svc.ServiceContext) cron.HandlerFunc {
    return func(ctx context.Context, t *cron.Task) error {
        var emails []types.EmailPayload
        if err := json.Unmarshal(t.Payload, &emails); err != nil {
            return err
        }
        l := batchLogic.NewBatchSendEmailLogic(ctx, svcCtx)
        return l.BatchSendEmail(emails)
    }
}
```

```go
// internal/handler/workers.go
serverCtx.CronServer.Add("batch_send_email", batch.BatchSendEmailHandler(serverCtx))
```

#### Notes

- Without passing `WithGroupAggregator`, the grouping feature does not work at all, and Group-related configuration parameters are ignored
- All client Push methods support the `asynq.Group()` option
- One Server has only one aggregator; multiple groups are routed via the `group` parameter

### Distributed Tracing

- Producer: In CommonClient, `otel.Inject` pushes TraceID into the Task header.
- Consumer: `TraceMiddleware` calls `otel.Extract` to restore the context.
- Result: you can see the complete timing diagram from API request to asynchronous task execution in Jaeger or Grafana Tempo.

Note: this tracing is integrated into the go-zero framework; you need to enable tracing in the go-zero project.

### Logger Replacement

By default, asynq's built-in logger is used. Replace it via `WithServerLogger`:

```go
cron := cron.MustNewServer(c.WorkConf, cron.WithServerLogger(&cron.AsynqLogger{}))
```

### Redis TLS

Configure TLS connections via `WithServerTLS` / `WithClientTLS`:

```go
tlsCfg := &tls.Config{InsecureSkipVerify: true}
server := cron.MustNewServer(conf, cron.WithServerTLS(tlsCfg))
client := cron.MustNewClient(conf, cron.WithClientTLS(tlsCfg))
```

### Metrics

Metrics are integrated into the go-zero Prometheus system and exposed via the `/metrics` endpoint.

#### Server Metrics (`cron_server_`)

##### Task Processing Metrics (Interceptor Collection)

| Metric | Type | Labels | Description |
|------|------|------|------|
| `cron_server_consume_total` | Counter | task_type, status | Consumption count (status: success/fail/skip_retry) |
| `cron_server_consume_duration_ms` | Histogram | task_type | Consumption duration |
| `cron_server_consume_bytes` | Counter | task_type | Consumption bytes |
| `cron_server_active_workers` | Gauge | task_type | Current concurrency |
| `cron_server_retry_total` | Counter | task_type | Retry execution count |
| `cron_server_skip_retry_total` | Counter | task_type | Skip retry count |
| `cron_server_panic_total` | Counter | task_type | Panic count (panics are not retried) |

##### Scheduler Metrics

| Metric | Type | Labels | Description |
|------|------|------|------|
| `cron_server_scheduler_trigger_total` | Counter | task_type | Scheduled task trigger count |
| `cron_server_scheduler_registered` | Gauge | - | Number of currently registered scheduled tasks |

##### Queue State Metrics (Collector Collection)

| Metric | Type | Labels | Description |
|------|------|------|------|
| `cron_server_tasks_enqueued_total` | Gauge | queue, state | Number of tasks in each state (state: active/pending/scheduled/retry/archived/completed) |
| `cron_server_queue_size` | Gauge | queue | Total number of tasks in the queue |
| `cron_server_queue_latency_seconds` | Gauge | queue | Queue latency (wait time of the oldest pending task) |
| `cron_server_queue_memory_usage_approx_bytes` | Gauge | queue | Queue memory usage (sampled estimate) |
| `cron_server_tasks_processed_total` | Counter | queue | Total number of processed tasks (success and failure) |
| `cron_server_tasks_failed_total` | Counter | queue | Total number of failed tasks |
| `cron_server_queue_paused_total` | Gauge | queue | Queue paused state |
| `cron_server_queue_groups` | Gauge | queue | Number of aggregation groups |
| `cron_server_tasks_aggregating_total` | Gauge | queue | Number of aggregating tasks |

> Queue state metrics are collected via a custom `QueueMetricsCollector`, supporting queue whitelist filtering to avoid metric mixing when multiple services share Redis.

#### Client Metrics (`cron_client_`)

| Metric | Type | Labels | Description |
|------|------|------|------|
| `cron_client_push_total` | Counter | task_type, push_type, status | Push count (push_type: immediate/delayed/scheduled) |
| `cron_client_push_duration_ms` | Histogram | task_type | Push duration |
| `cron_client_push_bytes` | Counter | task_type | Push bytes |
| `cron_client_cancel_total` | Counter | task_type, status | Cancel task count |

Note: these metrics require enabling Prometheus monitoring in the go-zero project.

## Full Example

### Using in go-zero

#### Server

**Directory Structure**

```shell
├── etc/
│   └── etc.yaml
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── handler/
│   │   ├── demoA/
│   │   │   └── gdemoahandler.go
│   │   └── workers.go
│   ├── logic/
│   │   └── demoA/
│   │       └── gdemoalogic.go
│   ├── svc/
│   │   └── servicecontext.go
│   └── types/
│       └── types.go
└── main.go
```

**Code**

```go
// internal/config/config.go
type BaseConfig struct {
    rest.RestConf
    WorkConf cron.ServerConfig
}
```

```go
// main.go
func main() {
    flag.Parse()

    var c config.BaseConfig
    conf.MustLoad(*configFile, &c, conf.UseEnv())
    ctx := svc.NewServiceContext(c)
    serviceGroup := service.NewServiceGroup()
    defer serviceGroup.Stop()
    handler.RegisterHandlers(serviceGroup, ctx)

    fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
    serviceGroup.Start()
}
```

```go
// internal/svc/servicecontext.go
type ServiceContext struct {
    Config     config.Config
    CronServer cron.Server
}

func NewServiceContext(c config.Config) *ServiceContext {
    c.WorkConf.Namespace = c.Name
    cronServer := cron.MustNewServer(c.WorkConf)

    return &ServiceContext{
        Config:     c,
        CronServer: cronServer,
    }
}
```

```go
// internal/handler/workers.go
import (
    "time"
    "github.com/hibiken/asynq"
)

func RegisterHandlers(server *service.ServiceGroup, serverCtx *svc.ServiceContext) {
    // 复用 go-zero RestConf.Timeout 接管定时任务超时（参见进阶指南：任务超时控制）
    var taskOpts []asynq.Option
    taskOpts = append(taskOpts,
        asynq.Timeout(time.Duration(serverCtx.Config.Timeout)*time.Millisecond),
        asynq.MaxRetry(0),
    )

    // 定时任务：一步完成调度注册 + handler 注册
    serverCtx.CronServer.CronAdd("*/1 * * * *", "GDemoA",
        demoA.GDemoAHandler(serverCtx), taskOpts...)

    server.Add(serverCtx.CronServer)
}
```

```go
// internal/handler/demoA/gdemoahandler.go
func GDemoAHandler(svcCtx *svc.ServiceContext) cron.HandlerFunc {
    return func(ctx context.Context, t *cron.Task) error {
        var req types.Name
        if err := json.Unmarshal(t.Payload, &req); err != nil {
            return err
        }
        l := demoA.NewGDemoALogic(ctx, svcCtx)
        return l.GDemoA(&req)
    }
}
```

```go
// internal/logic/demoA/gdemoalogic.go
type GDemoALogic struct {
    logx.Logger
    ctx    context.Context
    svcCtx *svc.ServiceContext
}

func NewGDemoALogic(ctx context.Context, svcCtx *svc.ServiceContext) *GDemoALogic {
    return &GDemoALogic{
        Logger: logx.WithContext(ctx),
        ctx:    ctx,
        svcCtx: svcCtx,
    }
}

func (l *GDemoALogic) GDemoA(req *types.Name) error {
    logc.Infof(l.ctx, "GDemoA called, %v", req)
    return nil
}
```

#### Client

**Important: when pushing tasks from the client, the taskType must match the full pattern registered on the server (including the Namespace prefix).**

```go
// internal/svc/servicecontext.go
type ServiceContext struct {
    Config     config.Config
    CronClient cron.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
    cronClient := cron.MustNewClient(c.ClientConf)

    return &ServiceContext{
        Config:     c,
        CronClient: cronClient,
    }
}
```

```go
// internal/logic/order_logic.go
import "github.com/hibiken/asynq"

func (l *OrderLogic) CreateOrder(req *types.OrderReq) error {
    payload := map[string]string{"email": "user@example.com", "content": "Welcome!"}

    // taskType 必须带 Namespace 前缀，Queue 必须指定为 Namespace
    namespace := l.svcCtx.Config.Name
    taskType := fmt.Sprintf("%s:%s", namespace, "send_email")
    _, err := l.svcCtx.CronClient.PushJson(l.ctx, taskType, payload, asynq.Queue(namespace))
    return err
}
```

```go
// 推荐：定义常量统一管理 taskType
const (
    Namespace     = "order-service"
    TaskSendEmail = Namespace + ":send_email"
    TaskSyncData  = Namespace + ":sync_data"
)

func (l *OrderLogic) CreateOrder(req *types.OrderReq) error {
    payload := map[string]string{"email": "user@example.com"}
    _, err := l.svcCtx.CronClient.PushJson(l.ctx, TaskSendEmail, payload, asynq.Queue(Namespace))
    return err
}
```

### Using in Standalone Scripts

#### Server

```go
func main() {
    conf := cron.ServerConfig{
        RedisConf:   cron.RedisConf{Addr: "localhost:6379", Mode: "single"},
        Concurrency: 10,
    }
    srv := cron.MustNewServer(conf)

    srv.Add("sync_data", func(ctx context.Context, t *cron.Task) error {
        fmt.Println("正在处理同步...")
        return nil
    })

    srv.Start()
}
```

#### Client

```go
func main() {
    clientConf := cron.ClientConfig{
        RedisConf: cron.RedisConf{Addr: "localhost:6379", Mode: "single"},
    }

    client := cron.MustNewClient(clientConf)
    defer client.Close()

    // 若服务端配置了 Namespace，taskType 必须带前缀，并指定 Queue
    client.Push(context.Background(), "data-service:raw_task", []byte("hello"), asynq.Queue("data-service"))

    // 若服务端未配置 Namespace，直接使用原始 pattern
    // client.Push(context.Background(), "raw_task", []byte("hello"))
}
```

## Changelog

See [CHANGELOG.md](./CHANGELOG.md)
