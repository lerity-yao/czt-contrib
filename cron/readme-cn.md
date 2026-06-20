# cron

[English](./README.md)

基于 [Asynq](https://github.com/hibiken/asynq) 构建的分布式任务队列系统，专为 Go-Zero 框架设计的定时任务和异步任务处理模块。

## 特性

- 🚀 **高性能**: 基于 Redis 的高性能分布式任务队列
- ⏰ **定时任务**: 支持 Cron 表达式定时任务
- 🔄 **异步处理**: 异步任务队列，支持延迟执行
- 📊 **监控指标**: 内置 Prometheus 指标收集
- 🔍 **链路追踪**: 集成 OpenTelemetry 链路追踪
- 🛡️ **错误恢复**: 自动 panic 恢复和错误处理
- 🔧 **配置灵活**: 支持多种 Redis 模式（单机、哨兵、集群）

## 安装

```bash
go get github.com/lerity-yao/czt-contrib/cron
```
## 服务生成

支持类似 goctl 工具一键生成服务端代码， 工具为 cztctl, 是 goctl 魔改的

也可以自定义服务端代码生成模板

请参考 [cztctl](https://github.com/lerity-yao/czt-contrib/blob/main/cztctl/README.md)

## ⚙️ 配置参数

### RedisConf (基础连接配置)
该配置控制如何连接到 Redis，支持 单机(Single)、哨兵(Sentinel) 和 集群(Cluster)。

| 参数名 | 类型 | 默认值 | 详细说明与建议 |
| --- | --- | --- | --- |
| Mode | string | single | 必填。可选：single, sentinel, cluster。决定了后续哪些字段生效。 |
| Addr | string | - | Mode=single 时必填。格式 "host:port"。 |
| Addrs | []string | - | Mode=cluster 时必填。集群种子节点列表，只需填入部分节点，驱动会自动发现全量拓扑。 |
| MasterName | string | - | Mode=sentinel 时必填。哨兵模式下监控的主节点名称（通常默认为 mymaster）。 |
| SentinelAddrs | []string | - | Mode=sentinel 时必填。哨兵节点列表。建议至少 3 个以保证高可用。 |
| Username | string | - | Redis 6.0+ ACL 认证用户名。 |
| Password | string | - | Redis 认证密码。 |
| DB | int64 | 0 | Redis 数据库索引。注意：Cluster 模式下此项无效。 |
| PoolSize | int64 | - | 连接池最大连接数。默认值为 10 * CPU核心数。高并发任务建议根据并发数调大。 |
| DialTimeout | int64 | 5 | 连接建立超时（秒）。网络环境差时可适当调大。 |
| ReadTimeout | int64 | 3 | 读超时（秒）。建议保留默认值。 |
| WriteTimeout | int64 | 3 | 写超时（秒）。建议保留默认值。 |

### ServerConfig (任务处理引擎配置)

该配置直接影响消费者的处理效率、稳定性和资源占用。

- Namespace (string):
核心逻辑：所有 Key 在 Redis 中都会加上此命名前缀。
建议：每个独立服务使用不同的 Namespace。这实现了物理隔离，防止不同服务的 Worker 误消费对方的任务。


- Concurrency (int64):
默认值：0（表示自动设置为 CPU 核心数）。
建议：如果任务涉及大量网络 IO（如发短信、请求第三方 API），建议调大至 20~100；如果是 CPU 密集型计算，建议保持默认或小幅调大。


- Queues (map[string]int):
核心逻辑：定义监听哪些队列及其权重。
实战举例：{"critical": 6, "default": 3, "low": 1} 表示 60% 的精力处理核心任务。

- StrictPriority (bool):
逻辑：若为 true，只要 critical 队列有一个任务，Worker 绝不会去碰 default 队列。
注意：开启此项可能导致低优先级队列“饥饿”（永远得不到处理），请谨慎使用。

- TaskCheckInterval (int64):
逻辑：所有队列都为空时，Worker 歇多久再去检查 Redis。
建议：默认 1 秒。过短会增加 Redis CPU 负担，过长会导致任务处理有明显延迟。


- ShutdownTimeout (int64):
逻辑：优雅停机时，Worker 等待当前任务完成的最长时间。
建议：默认 8 秒。如果你的任务逻辑很长（如处理大文件），必须调大此值，否则任务会被强行中断并重新入队。


- DelayedTaskCheckInterval (int64):
逻辑：检查“延时任务”和“重试任务”是否到点的频率。默认 5 秒。


- HealthCheckInterval (int64):
逻辑：Worker 与 Redis 的心跳检测。建议保持默认 15 秒。


- GroupGracePeriod (int64): 聚合窗口期（滑动窗口）。默认 60 秒。每次有新任务进入组时重置计时器，宽限期内无新任务才触发聚合。


- GroupMaxDelay (int64): 强制触发聚合的最长等待时间（硬上限）。默认 300 秒。从组内第一个任务到来算起，无论是否有新任务，到时间就强制聚合。防止 GracePeriod 被不断重置导致永不聚合。


- GroupMaxSize (int64): 组内任务达到多少个时，不等待窗口期直接触发聚合。默认 0（无限制）。


> 以上三个条件为 **OR** 关系，满足任意一个即触发聚合。需配合 `WithGroupAggregator` 使用，否则分组功能不生效。


- JanitorInterval (int64): 检查并清理 Redis 中已完成、过期任务的时间间隔。


- JanitorBatchSize (int64): 每次清理操作删除的数量上限。默认 100。防止一次性删除过多导致 Redis 阻塞。

**配置建议**
- 必须设置 Namespace：这是多服务共存的基础。
- 合理设置 Concurrency：IO 多则大，CPU 多则小。
- 设置 ShutdownTimeout：必须大于你业务逻辑中可能出现的最长耗时。

### ClientConfig (客户端配置)

仅包含 `RedisConf`，与 Server 共用同一套 Redis 连接配置。

### ServerOption

| Option | 参数 | 说明 |
|--------|------|------|
| `WithServerTLS` | `*tls.Config` | 设置 Redis TLS 连接配置 |
| `WithServerLogger` | `asynq.Logger` | 替换默认日志器（推荐 `&cron.AsynqLogger{}` 对接 go-zero logx） |
| `WithGroupAggregator` | `asynq.GroupAggregator` | 注入分组聚合器，启用 Group 功能 |
| `WithRetryDelayFunc` | `asynq.RetryDelayFunc` | 自定义重试退避策略（默认 `ExponentialRetryDelay`） |

### ClientOption

| Option | 参数 | 说明 |
|--------|------|------|
| `WithClientTLS` | `*tls.Config` | 设置 Redis TLS 连接配置 |


## API 参考

### 构造函数

| 函数 | 签名 | 说明 |
|------|------|------|
| `MustNewServer` | `func MustNewServer(conf ServerConfig, opts ...ServerOption) Server` | 创建 Server，失败 panic |
| `NewServer` | `func NewServer(conf ServerConfig, opts ...ServerOption) (Server, error)` | 创建 Server，失败返回 error |
| `MustNewClient` | `func MustNewClient(conf ClientConfig, opts ...ClientOption) *CommonClient` | 创建 Client，失败 panic |
| `NewClient` | `func NewClient(conf ClientConfig, opts ...ClientOption) (*CommonClient, error)` | 创建 Client，失败返回 error |
| `MustNewClientFromRedisClient` | `func MustNewClientFromRedisClient(rds redis.UniversalClient) *CommonClient` | 从已有 Redis 连接创建 Client，失败 panic |
| `NewClientFromRedisClient` | `func NewClientFromRedisClient(rds redis.UniversalClient) (*CommonClient, error)` | 从已有 Redis 连接创建 Client，失败返回 error |

### 公开类型

| 类型 | 定义 | 说明 |
|------|------|------|
| `Task` | `struct { Type string; Payload []byte }` | Handler 接收的任务载体，Type 为任务类型，Payload 为原始字节数据 |
| `HandlerFunc` | `func(ctx context.Context, t *Task) error` | 任务处理函数签名，所有 Handler 均需实现此类型 |
| `AsynqLogger` | `struct{}` | 内置日志适配器，将 asynq 日志桥接到 go-zero logx |

### Server 接口方法

| 方法 | 签名 | 说明 |
|------|------|------|
| `Add` | `Add(pattern string, handler HandlerFunc)` | 注册任务 Handler（消费外部投递的任务），自动拼接 Namespace 前缀 |
| `CronAdd` | `CronAdd(spec string, pattern string, handler HandlerFunc, opts ...asynq.Option) string` | 一步注册定时任务（自产自销），同时完成 handler 注册与定时调度，自动设置 TaskID 去重，返回 EntryID |
| `SetBaseContext` | `SetBaseContext(ctx context.Context)` | 注入基础上下文，所有任务 handler 的 ctx 以此为父级，必须在 Start() 之前调用 |
| `Start` | `Start()` | 启动 Scheduler + Processor，非阻塞 |
| `Stop` | `Stop()` | 优雅停机：Scheduler → Server → Inspector 顺序关闭 |

### Client 接口方法

| 方法 | 签名 | 说明 |
|------|------|------|
| `Push` | `Push(ctx, taskType string, payload []byte, opts ...asynq.Option) (*asynq.TaskInfo, error)` | 立即执行，推送原始字节 |
| `PushJson` | `PushJson(ctx, taskType string, data any, opts ...asynq.Option) (*asynq.TaskInfo, error)` | 立即执行，自动 JSON 序列化 |
| `PushIn` | `PushIn(ctx, taskType string, payload []byte, delay time.Duration, opts ...) (*asynq.TaskInfo, error)` | 延时执行，指定 Duration |
| `PushInJson` | `PushInJson(ctx, taskType string, data any, delay time.Duration, opts ...) (*asynq.TaskInfo, error)` | 延时执行，自动 JSON 序列化 |
| `PushAt` | `PushAt(ctx, taskType string, payload []byte, at time.Time, opts ...) (*asynq.TaskInfo, error)` | 定点执行，指定绝对时间 |
| `PushAtJson` | `PushAtJson(ctx, taskType string, data any, at time.Time, opts ...) (*asynq.TaskInfo, error)` | 定点执行，自动 JSON 序列化 |
| `CancelTask` | `CancelTask(queue, taskID string) error` | 撤回 Scheduled/Pending/Retry 状态的任务 |
| `RescheduleTask` | `RescheduleTask(ctx, queue, taskID, taskType string, data any, newDelay time.Duration, opts ...) (*asynq.TaskInfo, error)` | 原子撤回 + 重新投递，TaskID 不变 |
| `Close` | `Close() error` | 关闭客户端连接 |

> 所有 Push 系列方法默认 `MaxRetry=0`（不重试），可通过 opts 覆盖。

### 工具函数

| 函数 | 签名 | 说明 |
|------|------|------|
| `ExponentialRetryDelay` | `func(n int, _ error, _ *asynq.Task) time.Duration` | 内置指数退避（`2^n - 1` 秒），Server 默认使用 |
| `GetTaskID` | `func(ctx context.Context) (string, bool)` | 任务 ID，重试过程中保持不变 |
| `GetRetryCount` | `func(ctx context.Context) (int, bool)` | 当前重试次数，0 = 首次执行 |
| `GetMaxRetry` | `func(ctx context.Context) (int, bool)` | 最大重试次数上限 |
| `GetQueueName` | `func(ctx context.Context) (string, bool)` | 任务所在队列名 |

## 进阶指南

### Namespace 与 TaskType 匹配规则

**这是一个重要的注意事项，客户端投递任务时必须了解服务端的 Namespace 配置。**

服务端配置 `Namespace` 后会产生两个影响：

1. **taskType 自动拼接**：`Add` 方法将 pattern 转换为 `Namespace:pattern` 格式
2. **队列自动创建**：服务端监听以 Namespace 命名的队列

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

**在 go-zero 环境中**，通常会将服务名作为 Namespace：

```go
// serviceContext.go 中常见写法
c.WorkConf.Namespace = c.Name  // 使用 go-zero 配置的服务名
```

**客户端投递任务时必须同时满足两个条件：**

1. **taskType 带 Namespace 前缀**：与服务端注册的完整 pattern 一致
2. **指定队列**：通过 `asynq.Queue(namespace)` 指定任务投递到哪个队列

```go
import "github.com/hibiken/asynq"

// 假设服务端配置：Namespace = "order-service"，注册：Add("send_email", handler)
// 服务端实际注册的 pattern："order-service:send_email"
// 服务端监听的队列："order-service"

// ✔️ 正确写法：taskType 带 namespace + 指定队列
client.PushJson(ctx, "order-service:send_email", payload, asynq.Queue("order-service"))
```

**总结**：
- 客户端和服务端统一 Namespace 命名规范
- 定义常量统一管理 taskType 和 Namespace，避免硬编码
- 若服务端未配置 Namespace，则 taskType 直接使用原始 pattern，无需指定队列（默认 "default"）

### 重试退避策略

Server 默认使用 `ExponentialRetryDelay` 指数退避策略，延迟公式为 `2^n - 1` 秒：

| 重试次数 | 延迟 |
|---------|------|
| 1 | 1s |
| 2 | 3s |
| 3 | 7s |
| 4 | 15s |
| 5 | 31s |
| 6 | 63s（~1分钟） |
| 7 | 127s（~2分钟） |
| 10 | 1023s（~17分钟） |

如需自定义退避策略，可通过 `WithRetryDelayFunc` 覆盖：

```go
server := cron.MustNewServer(c.WorkConf,
    cron.WithRetryDelayFunc(func(n int, e error, t *asynq.Task) time.Duration {
        // 固定间隔 10 秒
        return 10 * time.Second
    }),
)
```

#### 任务元信息

在 Handler 中可通过 `ctx` 获取任务运行时元信息：

```go
retryCount, _ := cron.GetRetryCount(ctx)  // 当前第几次重试（0 = 首次执行）
maxRetry, _   := cron.GetMaxRetry(ctx)    // 最大重试次数
taskID, _     := cron.GetTaskID(ctx)      // 任务 ID（重试不变）
queueName, _  := cron.GetQueueName(ctx)   // 队列名
```

#### 放弃重试

当错误不可恢复时（参数非法、业务拒绝、幂等冲突），使用 `asynq.SkipRetry` 放弃重试：

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

> **判断标准：** 重试了结果还是一样 → SkipRetry；重试了有可能成功 → 让它重试。
>
> **注意：** 必须使用 `fmt.Errorf("%w", asynq.SkipRetry)` 包装。`github.com/pkg/errors` 的 `errors.Wrap` 不兼容标准库 `errors.Is()` 解包链，asynq 无法识别。

### 任务超时控制

> **v0.1.0 起**：定时任务与投递任务的执行超时均可由业务侧 **按任务粒度** 注入，handler 收到的 `ctx` 自带 `Deadline`，到期后 `ctx.Done()` 自动触发，业务可主动退出。

#### 设计要点

- asynq 默认超时为硬编码常量 **30 分钟**，且无法通过 `asynq.Config` 全局配置。
- 单任务超时通过入队元数据 `asynq.Timeout(d)` 写入 `msg.Timeout`，processor 拉出任务时调用 `context.WithDeadline(baseCtx, deadline)` 注入 ctx。
- 仅入队动作（`CronAdd`、`Client.Push*`）能携带 `asynq.Timeout`；纯 handler 注册的 `Add` 不参与投递，无法设置。

#### 定时任务：用 go-zero RestConf.Timeout 接管

`rest.RestConf.Timeout` 默认 `3000ms`、必 > 0，可在 `workers.go` 直接复用为定时任务超时基准：

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

效果：定时到点入队时 `msg.Timeout` 自动写入 yaml 配置的 `Timeout`，handler 收到的 ctx `Deadline = now + Timeout`。改 yaml 即可全局调整所有定时任务的超时上限，无需改代码。

> 默认 3 秒对部分 IO 密集型定时任务偏短，建议在 yaml 中显式配置（如 `Timeout: 30000`）。

#### 投递任务：客户端按任务自定义超时

`Client.Push*` 系列均支持透传 `asynq.Timeout(d)`，由投递方按任务实际耗时按需指定，**不受 30 分钟限制**：

```go
import "github.com/hibiken/asynq"

// 短任务：5 秒超时
client.PushJson(ctx, "order-service:send_sms", payload,
    asynq.Queue("order-service"),
    asynq.Timeout(5*time.Second))

// 长任务：1 小时超时
client.PushJson(ctx, "order-service:export_report", payload,
    asynq.Queue("order-service"),
    asynq.Timeout(time.Hour))
```

> 投递方不传 `asynq.Timeout` 时回落到 asynq 默认 30 分钟。生产环境建议每个任务按耗时预算显式设置。

### 任务分组聚合（Group Aggregation）

将多个同组任务合并为一个批量任务再处理，适用于通知合并、批量写入等场景。

#### 触发条件（三选一）

| 条件 | 配置参数 | 默认值 | 触发机制 |
|------|----------|--------|----------|
| 宽限期超时 | `GroupGracePeriod` | 60s | **滑动窗口**：每来一个新任务重置计时器，宽限期内无新任务才触发 |
| 最大延迟 | `GroupMaxDelay` | 300s | **硬上限**：从第一个任务到来算起，到时间强制聚合（兜底） |
| 数量达标 | `GroupMaxSize` | 0（无限制） | 组内任务数达到阈值，立即聚合 |

#### 触发场景示例

假设配置：`GracePeriod=60s, MaxDelay=300s, MaxSize=50`

| 场景 | 行为 |
|------|------|
| 来了 3 个任务后不再来 | 最后一个任务到达 60s 后聚合（命中 GracePeriod） |
| 每隔 30s 持续来任务 | 到 300s 时强制聚合（命中 MaxDelay 兜底） |
| 短时间连续来了 50 个任务 | 立即聚合（命中 MaxSize） |
| 每隔 61s 来一个任务 | 每个任务单独聚合（61s > 60s，GracePeriod 先触发） |

#### 使用方式

**1. 服务端注入聚合器（必需）**

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

> 整个 Server 只有一个 `GroupAggregator`，通过 `group` 参数区分不同分组的聚合逻辑。

**2. 客户端投递时指定分组**

```go
import "github.com/hibiken/asynq"

// 投递到 "batch_email" 分组
client.PushJson(ctx, "order-service:send_email", payload,
    asynq.Queue("order-service"),
    asynq.Group("batch_email"),  // 指定分组名
)
```

**3. 服务端注册合并后任务的 Handler**

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

#### 注意事项

- 不传 `WithGroupAggregator` 时，分组功能完全不生效，Group 相关配置参数为摆设
- 客户端所有 Push 系列方法均支持 `asynq.Group()` 选项
- 一个 Server 只有一个聚合器，多个分组通过 `group` 参数路由

### 链路追踪

- 生产者：在 CommonClient 中通过 otel.Inject 将 TraceID 压入 Task 的 Header。
- 消费者：通过 TraceMiddleware 调用 otel.Extract 恢复上下文。
- 结果：你可以在 Jaeger 或 Grafana Tempo 中看到从 API 请求到异步任务执行的完整时序图。

注意，这链路跟踪是集成在 go-zero 框架中的，你需要在 go-zero 项目中开启链路跟踪功能。

### 日志替换

默认使用 asynq 自带日志。通过 `WithServerLogger` 替换：

```go
cron := cron.MustNewServer(c.WorkConf, cron.WithServerLogger(&cron.AsynqLogger{}))
```

### Redis TLS

通过 `WithServerTLS` / `WithClientTLS` 配置 TLS 连接：

```go
tlsCfg := &tls.Config{InsecureSkipVerify: true}
server := cron.MustNewServer(conf, cron.WithServerTLS(tlsCfg))
client := cron.MustNewClient(conf, cron.WithClientTLS(tlsCfg))
```

### 监控指标

监控指标已并入 go-zero 的 Prometheus 体系，通过 `/metrics` 端点暴露。

#### Server 端指标 (`cron_server_`)

##### 任务处理指标（Interceptor 采集）

| 指标 | 类型 | 标签 | 说明 |
|------|------|------|------|
| `cron_server_consume_total` | Counter | task_type, status | 消费计数（status: success/fail/skip_retry） |
| `cron_server_consume_duration_ms` | Histogram | task_type | 消费耗时 |
| `cron_server_consume_bytes` | Counter | task_type | 消费字节数 |
| `cron_server_active_workers` | Gauge | task_type | 当前并发数 |
| `cron_server_retry_total` | Counter | task_type | 重试执行次数 |
| `cron_server_skip_retry_total` | Counter | task_type | 跳过重试次数 |
| `cron_server_panic_total` | Counter | task_type | panic 次数（panic 不重试） |

##### Scheduler 指标

| 指标 | 类型 | 标签 | 说明 |
|------|------|------|------|
| `cron_server_scheduler_trigger_total` | Counter | task_type | 定时任务触发次数 |
| `cron_server_scheduler_registered` | Gauge | - | 当前注册的定时任务数 |

##### 队列状态指标（Collector 采集）

| 指标 | 类型 | 标签 | 说明 |
|------|------|------|------|
| `cron_server_tasks_enqueued_total` | Gauge | queue, state | 各状态任务数（state: active/pending/scheduled/retry/archived/completed） |
| `cron_server_queue_size` | Gauge | queue | 队列任务总数 |
| `cron_server_queue_latency_seconds` | Gauge | queue | 队列延迟（最旧 pending 任务等待时间） |
| `cron_server_queue_memory_usage_approx_bytes` | Gauge | queue | 队列内存占用（采样估算值） |
| `cron_server_tasks_processed_total` | Counter | queue | 已处理任务总数（含成功和失败） |
| `cron_server_tasks_failed_total` | Counter | queue | 失败任务总数 |
| `cron_server_queue_paused_total` | Gauge | queue | 队列暂停状态 |
| `cron_server_queue_groups` | Gauge | queue | 聚合组数量 |
| `cron_server_tasks_aggregating_total` | Gauge | queue | 聚合中的任务数 |

> 队列状态指标通过自定义 `QueueMetricsCollector` 采集，支持队列白名单过滤，解决多服务共用 Redis 时指标混杂问题。

#### Client 端指标 (`cron_client_`)

| 指标 | 类型 | 标签 | 说明 |
|------|------|------|------|
| `cron_client_push_total` | Counter | task_type, push_type, status | 投递计数（push_type: immediate/delayed/scheduled） |
| `cron_client_push_duration_ms` | Histogram | task_type | 投递耗时 |
| `cron_client_push_bytes` | Counter | task_type | 投递字节数 |
| `cron_client_cancel_total` | Counter | task_type, status | 撤销任务计数 |

注意：这些指标需要在 go-zero 项目中开启 Prometheus 监控功能。

## 完整示例

### 在 go-zero 中使用

#### Server

**目录结构**

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

**代码**

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

**重要：客户端投递任务时，taskType 必须与服务端注册的完整 pattern 一致（包含 Namespace 前缀）。**

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

### 独立脚本中使用

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

## 更新日志

查看 [CHANGELOG.md](./CHANGELOG.md)
