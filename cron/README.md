# cron

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


- GroupGracePeriod (int64): 聚合窗口期。默认 60 秒。即第一个任务进入组后，等多久才触发聚合。


- GroupMaxDelay (int64): 强制触发聚合的最长等待时间。


- GroupMaxSize (int64): 组内任务达到多少个时，不等待窗口期直接触发聚合。


- JanitorInterval (int64): 检查并清理 Redis 中已完成、过期任务的时间间隔。


- JanitorBatchSize (int64): 每次清理操作删除的数量上限。默认 100。防止一次性删除过多导致 Redis 阻塞。

**配置建议**
- 必须设置 Namespace：这是多服务共存的基础。
- 合理设置 Concurrency：IO 多则大，CPU 多则小。
- 设置 ShutdownTimeout：必须大于你业务逻辑中可能出现的最长耗时。


## 💎 核心接口能力详解

### Server 接口：高性能消费者与调度引擎

Server 封装了任务的获取、解码、中间件执行及定时触发逻辑。

| 接口方法 | 参数说明                                                                      | 核心能力                                                                                                                                  |
| -------- |---------------------------------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------|
| Add      | pattern: 任务类型 <br>handler: 处理函数<br>cronSpec: Cron 表达式<br>opts: Asynq 原生配置 | 三位一体注册：<br>1. 自动拼接 Namespace:Pattern。<br>2. 逻辑隔离：防止不同服务误消费。<br>3. 自产自销：若有 cronSpec 则自动注册为定时任务，否则作为普通 Worker。<br> 普通的work需要有client投递任务 |
| Start    | 无                                                                         | 异步启动：启动 Scheduler 和 Processor 后立即返回。适用于 go-zero 的 ServiceGroup 管理，不会阻塞主线程。                                                            |
| Stop     | 无                                                                         | 优雅停机：按照 Scheduler -> Server -> Inspector 顺序关闭。先停产，再清空存量任务，最后释放 Redis 连接。                                                              |
| CronAdd | spec: Cron 表达式<br>pattern: 任务类型<br>opts: Asynq 原生配置 | 注册定时任务：根据 Cron 表达式自动触发任务。支持秒级精度。 |
### Client 接口：强类型生产者与任务控制器
Client 提供了多种任务进入 Redis 的姿势。

| 接口方法 | 类型 | 核心能力 |
| -------- | ---- | -------- |
| Push / PushJson | 立即 | 支持 any 类型自动 JSON 序列化，注入 TraceID 后推入队列。 |
| PushIn / PushInJson | 延时 | 允许指定 Duration（如 1h 后执行）。常用于延迟补偿、超时处理。 |
| PushAt / PushAtJson | 定时 | 指定绝对时间点（time.Time）。 |
| CancelTask | 控制 | 根据 TaskID 撤回处于 Scheduled (延时)、Pending (排队) 状态的任务。 |
| RescheduleTask | 控制 | 原子化实现“撤回 + 重新按新延迟投递”。支持固定 TaskID 确保幂等性。 |

**定时循环执行任务不支持投递，只能在server端注册，server端会根据cronSpec注册定时触发任务**

## 链路跟踪

- 生产者：在 CommonClient 中通过 otel.Inject 将 TraceID 压入 Task 的 Header。
- 消费者：通过 TraceMiddleware 调用 otel.Extract 恢复上下文。
- 结果：你可以在 Jaeger 或 Grafana Tempo 中看到从 API 请求到异步任务执行的完整时序图。

注意，这链路跟踪是集成在 go-zero 框架中的，你需要在 go-zero 项目中开启链路跟踪功能。

## 监控指标

监控指标已并入 go-zero 的 Prometheus 体系，通过 `/metrics` 端点暴露。

### Server 端指标 (`cron_server_`)

#### 任务处理指标（Interceptor 采集）

| 指标 | 类型 | 标签 | 说明 |
|------|------|------|------|
| `cron_server_consume_total` | Counter | task_type, status | 消费计数（status: success/fail/skip_retry） |
| `cron_server_consume_duration_ms` | Histogram | task_type | 消费耗时 |
| `cron_server_consume_bytes` | Counter | task_type | 消费字节数 |
| `cron_server_active_workers` | Gauge | task_type | 当前并发数 |
| `cron_server_retry_total` | Counter | task_type | 重试执行次数 |
| `cron_server_skip_retry_total` | Counter | task_type | 跳过重试次数 |
| `cron_server_panic_total` | Counter | task_type | panic 次数（panic 不重试） |

#### Scheduler 指标

| 指标 | 类型 | 标签 | 说明 |
|------|------|------|------|
| `cron_server_scheduler_trigger_total` | Counter | task_type | 定时任务触发次数 |
| `cron_server_scheduler_registered` | Gauge | - | 当前注册的定时任务数 |

#### 队列状态指标（Collector 采集）

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

### Client 端指标 (`cron_client_`)

| 指标 | 类型 | 标签 | 说明 |
|------|------|------|------|
| `cron_client_push_total` | Counter | task_type, push_type, status | 投递计数（push_type: immediate/delayed/scheduled） |
| `cron_client_push_duration_ms` | Histogram | task_type | 投递耗时 |
| `cron_client_push_bytes` | Counter | task_type | 投递字节数 |
| `cron_client_cancel_total` | Counter | task_type, status | 撤销任务计数 |

注意：这些指标需要在 go-zero 项目中开启 Prometheus 监控功能。

## 日志

默认在日志中使用的是 asynq 自带的日志。你可以显性的通过调用 `WithServerLogger` 来指定日志器。

```go
// 使用 go-zero logx 替换 asynq 自带的日志
cron := cron.MustNewServer(c.WorkConf, cron.WithServerLogger(&cron.AsynqLogger{}))
```

## Redis TLS

可以显示的类型于 `WithServerLogger` 一样，通过 `WithServerTLS` 来指定 TLS 配置。

## server 使用
### 在 go-zero 中使用

#### 目录结构
```shell
├── etc
│   └── etc.yaml
├── go.mod
├── go.sum
├── internal
│   ├── config
│   │   └── config.go
│   ├── handler
│   │   ├── demoA
│   │   │   └── demoAhanadler.go
│   │   └── worker.go
│   ├── logic
│   │   └── demoA
│   │       └── demoAxxxLogic.go
│   ├── svc
│   │   └── serviceContext.go
│   └── types
│       └── types.go
└── main.go
```

#### 代码
```go
// internal/config/config.go
type BaseConfig struct {
    rest.RestConf
    WorkConf             cron.ServerConfig
}
```

```go
// main.go
var configFile = flag.String("f", "etc/etc.yaml", "the config file")

func main() {
    flag.Parse()
    
    // 加载基础配置
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
// internal/svc/serviceContext.go
type ServiceContext struct {
    Config config.Config
    Cron   cron.Server
}

func NewServiceContext(c config.Config) *ServiceContext {
    c.WorkConf.Namespace = c.Name
    cron := cron.MustNewServer(c.WorkConf, cron.WithServerLogger(&cron.AsynqLogger{}))
    
    return &ServiceContext{
        Config: c,
        Cron:   cron,
    }
}
```

```go
// internal/handler/worker.go
func RegisterHandlers(server *service.ServiceGroup, serverCtx *svc.ServiceContext) {
    serverCtx.Cron.Add("demoA", demoA.DemoAHandle(serverCtx))
    server.Add(serverCtx.Cron)
}
```

```go
// internal/handler/demoA/demoAhandler.go
// 定时任务，没有req，如果是其他的比如延时，指定时间，立即执行，需要把json部分代码注释去掉
func DemoAHandle(svcCtx *svc.ServiceContext) cron.HandlerFunc {
    return func(ctx context.Context, t *cron.Task) error {
        var req types.DemoAxxxReq
        //err := json.Unmarshal(t.Payload, &req)
        //if err != nil {
        //	return err
        //}
        l := demoA.NewDemoAxxxLogic(ctx, svcCtx)
        return l.NewDemoAxxx(req)
    }
}
```

```go
// internal/logic/demoA/demoAxxxLogic.go
package demoA

import (
	"context"
	"example/internal/svc"
	"example/internal/types"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type DemoAxxxLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDemoAxxxLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DemoAxxxLogic {
	return &DemoAxxxLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DemoAxxxLogic) NewDemoAxxx(req types.DemoAxxxReq) error {
	logc.Infof(l.ctx, "NewDemoAxxx called, %v", req)
	return nil
}
```

启动项目，就能看到日志

里面关于consul的，并没有在上面代码提现

```shell
API server listening at: 127.0.0.1:41349
{"@timestamp":"2026-02-08T13:55:44.236+08:00","caller":"devserver/server.go:71","content":"Starting dev http server at :6060","level":"info"}
register center consul url is 0.0.0.0:8886
 type is ttl
{"@timestamp":"2026-02-08T13:55:44.369+08:00","caller":"consul@v0.1.5/register.go:159","content":"Service tax-invoice.cron id tax-invoice.cron-192.168.13.72-8886 registered successfully","level":"info"}
{"@timestamp":"2026-02-08T13:55:44.369+08:00","caller":"cron/server.go:146","content":"[ASYNQ] Cron job registered: [*/1 * * * *] -> tax-invoice.cron:demoA (EntryID: 902bd196-a99f-427f-bc0b-f06419781d19)","level":"info"}
{"@timestamp":"2026-02-08T13:55:44.369+08:00","caller":"cron/server.go:132","content":"[ASYNQ] 注册定时任务: tax-invoice.cron:demoA","level":"info"}
Starting server at 0.0.0.0:8886...
{"@timestamp":"2026-02-08T13:55:44.369+08:00","caller":"cron/log.go:16","content":"Scheduler starting","level":"info"}
{"@timestamp":"2026-02-08T13:55:44.370+08:00","caller":"cron/log.go:16","content":"Scheduler timezone is set to Local","level":"info"}
{"@timestamp":"2026-02-08T13:55:44.370+08:00","caller":"cron/log.go:16","content":"Starting processing","level":"info"}
{"@timestamp":"2026-02-08T13:55:44.370+08:00","caller":"cron/log.go:16","content":"Send signal TSTP to stop processing new tasks","level":"info"}
{"@timestamp":"2026-02-08T13:55:44.370+08:00","caller":"cron/log.go:16","content":"Send signal TERM or INT to terminate the process","level":"info"}
```

执行 `curl http://127.0.0.1:6060/metrics` 可以看到 Prometheus 指标

#### 在独立脚本或非 go-zero 项目中使用

```go
func main() {
    conf := cron.ServerConfig{
        RedisConf: cron.RedisConf{Addr: "localhost:6379", Mode: "single"},
        Concurrency: 10,
    }
    srv := cron.MustNewServer(conf)
    
    srv.Add("sync_data", func(ctx context.Context, t *cron.Task) error {
        fmt.Println("正在处理同步...")
        return nil
    }, "")
	
    srv.Start()
}
```

## Client 使用

`Client` 不仅支持简单的任务发送，还深度集成了 **OpenTelemetry 链路追踪** 和 **任务生命周期控制**。

### go-zero中使用

```go
// svc 
type ServiceContext struct {
    Config      config.Config
	Cron        cron.Client // 定义 Client 接口 
}

func NewServiceContext(c config.Config) *ServiceContext {
    // 1. 初始化 Client (建议使用 MustNewClient 简化逻辑)
    // 支持通过 Option 注入 TLS
    cronClient := cron.MustNewClient(c.WorkConf.ClientConfig,
        cron.WithClientTLS(xxx)
	) // 如果有证书则传入 tls.Config
    
    return &ServiceContext{
        Config:      c,
        Cron:        cronClient,
    }
}
```   

```go
// logic
func (l *OrderLogic) CreateOrder(req *types.OrderReq) error {
    // 业务逻辑处理...
    
    // 异步投递：发送确认邮件
    // 优势：自动携带当前请求的 TraceID，实现全链路追踪
    payload := map[string]string{"email": "user@example.com", "content": "Welcome!"}
    
    _, err := l.svcCtx.Cron.PushJson(l.ctx, "send_email", payload)
    if err != nil {
        return err
    }
    
    return nil
}
```

### 独立脚本或非 go-zero 项目中使用

```go
func main() {
    clientConf := cron.ClientConfig{
        RedisConf: cron.RedisConf{Addr: "localhost:6379", Password: "xxx"},
    }
    
    // 支持通过 Option 注入 TLS
    client := cron.MustNewClient(clientConf, cron.WithClientTLS(myTlsConfig))
    defer client.Close()

    // 投递普通字节数据
    client.Push(context.Background(), "raw_task", []byte("hello world"))
}
```

## 更新日志

查看 [CHANGELOG.md](./CHANGELOG.md)







