# Changelog

所有版本变更记录。格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)。

## [0.1.1] - 2026-06-04

### 依赖升级

- `github.com/zeromicro/go-zero` v1.10.0 → v1.10.2
- `github.com/redis/go-redis/v9` v9.17.3 → v9.20.0
- 同步 `go mod tidy` 清理无用 indirect 依赖
- 主动保留 `go.opentelemetry.io/otel*` v1.40.0（v1.44+ 强制要求 go 1.25，为兼容使用方不升）
- `go` directive 保持 1.24.0

## [0.1.0] - 2026-06-04

### 破坏性变更 (Breaking)

- **`CronAdd` 签名调整**：新增必传参数 `handler HandlerFunc`，一步完成调度注册与 handler 注册，避免同一 task type 在 `CronAdd`/`Add` 两处重复书写
  - 旧用法：`CronAdd(spec, pattern, opts...)` + `Add(pattern, handler)`
  - 新用法：`CronAdd(spec, pattern, handler, opts...)`
- **定时任务超时改由 go-zero `RestConf.Timeout` 接管**：v0.1.0 之前定时任务未注入超时，行为回落到 asynq 默认 30 分钟硬编码常量；v0.1.0 起推荐在 `workers.go` 将 `serverCtx.Config.Timeout`（go-zero `rest.RestConf.Timeout`）通过 `asynq.Timeout(d)` 注入 `CronAdd`，使 handler 收到的 `ctx` 自带 `Deadline`，与 go-zero HTTP 超时管控统一
  - cztctl 模板将在后续版本同步联动，业务方升级时需手动调整 `workers.go`，详见迁移指引

### 迁移指引

**API 签名迁移**

```diff
- serverCtx.CronServer.CronAdd("*/1 * * * *", "GDemoB")
- serverCtx.CronServer.Add("GDemoB", demoA.GDemoBHandler(serverCtx))
+ serverCtx.CronServer.CronAdd("*/1 * * * *", "GDemoB", demoA.GDemoBHandler(serverCtx))
```

**接入 go-zero 超时管控**

```diff
+ import (
+     "time"
+     "github.com/hibiken/asynq"
+ )
+
  func RegisterHandlers(server *service.ServiceGroup, serverCtx *svc.ServiceContext) {
+     var taskOpts []asynq.Option
+     taskOpts = append(taskOpts,
+         asynq.Timeout(time.Duration(serverCtx.Config.Timeout)*time.Millisecond),
+         asynq.MaxRetry(0),
+     )
-     serverCtx.CronServer.CronAdd("*/1 * * * *", "GDemoB", demoA.GDemoBHandler(serverCtx))
+     serverCtx.CronServer.CronAdd("*/1 * * * *", "GDemoB",
+         demoA.GDemoBHandler(serverCtx), taskOpts...)
      server.Add(serverCtx.CronServer)
  }
```

> `rest.RestConf.Timeout` 默认 `3000ms`，对 IO 密集型定时任务偏短，建议在 yaml 中显式调大（如 `Timeout: 30000`）。

### 新增

- `CronAdd` / `Client.Push*` 可通过 `asynq.Timeout(d)` opt 为单个任务设置任意时长超时（不受 asynq 默认 30 分钟限制），例：`CronAdd("*/1 * * * *", "X", h, asynq.Timeout(time.Hour))`
- README 进阶指南新增「任务超时控制」章节，详述定时任务与投递任务的超时注入机制

### 修复

- `ExponentialRetryDelay` 公式修正为 `2^(n+1) - 1`：asynq 传入的 `n` 为上次已重试次数（0-indexed），原公式 `2^n - 1` 在第 1 次重试时返回 0s，导致首次重试无退避；修正后重试序列稳定为 `1s, 3s, 7s, 15s, 31s, ...`

## [0.0.9] - 2026-06-04

### 新增

- 新增 `SetBaseContext(ctx context.Context)` 方法，支持注入基础上下文，所有任务 handler 的 ctx 将以此为父级

### 变更

- `asynq.Server` 延迟到 `Start()` 时创建，支持在构造后、启动前动态配置上下文

## [0.0.8] - 2026-06-04

### 新增

- 新增 `GetTaskID`、`GetRetryCount`、`GetMaxRetry`、`GetQueueName` 包装函数，Handler 层无需直接依赖 asynq 即可获取任务元信息

### 变更

- README.md 全面重构：完整示例按 go-zero / 独立脚本 + Server / Client 四象限组织
- 进阶指南新增任务分组聚合、放弃重试用法示例
- 补充 `asynq.SkipRetry` 包装注意事项（`pkg/errors` 不兼容）
- 监控指标章节标题层级修正

## [0.0.7] - 2026-06-04

### 新增

- 新增 `WithGroupAggregator` ServerOption，支持注入任务分组聚合器，使 Group 分组功能可用
- 新增 `WithRetryDelayFunc` ServerOption，支持自定义重试延迟策略
- 新增 `ExponentialRetryDelay` 内置重试策略（`2^n - 1` 秒：1s, 3s, 7s, 15s, 31s...）
- `CronAdd` 自动以 `realPattern` 作为 TaskID，多实例部署时定时任务不会重复投递

### 修复

- 移除 `ServerConfig.Validate()` 中 `GroupGracePeriod` 的死代码校验（int64 不可能 >0 且 <1）

### 变更

- `GroupMaxDelay` 默认值从 0（无限制）调整为 300 秒（5分钟），防止任务持续到来时永不聚合
- Server 默认重试策略从 asynq 内置的 `n⁴` 指数退避改为 `ExponentialRetryDelay`（`2^n - 1` 秒）
- `Add` 日志统一为 `[CRON] Worker registered: %s`

## [0.0.6] - 2026-05-20

### 变更

- `metrics.go` 中 `log.Printf` 替换为 `logx.Errorf`，统一使用 go-zero 日志规范
- 移除 `LoggingMiddleware`，避免与 asynq 内部错误日志重复
- 移除 `RecoveryMiddleware` 中冗余的 `logc.Errorf`，asynq `perform` 已内置 panic recovery 日志

## [0.0.5] - 2026-05-15

### 新增

- 新增完整监控指标体系
  - Server 端：`cron_server_consume_total`、`cron_server_consume_duration_ms`、
   `cron_server_consume_bytes`、`cron_server_active_workers`、`cron_server_retry_total`、
   `cron_server_skip_retry_total`、`cron_server_panic_total`、`queue_groups`、
   `tasks_aggregating_total`、`cron_scheduler_trigger_total`、`cron_scheduler_registered`、
   `cron_tasks_enqueued_total`、`cron_queue_size`、`cron_queue_latency_seconds`
  - Client 端：`cron_client_push_total`、`cron_client_push_duration_ms`、`cron_client_push_bytes`、`cron_client_cancel_total`
- 新增队列白名单过滤，解决多服务共用 Redis 时指标混杂问题
- Client 默认不重试（MaxRetry=0），用户可通过 opts 覆盖

### 变更

- 监控指标命名空间从 `asynq_` 改为 `cron_`
- Server 端指标添加 `server` subsystem

## [0.0.4] - 2026-04-10

### 变更

- 升级 Go 版本至 1.24.0
- 升级 go-zero 至 v1.10.0
- 更新依赖版本
