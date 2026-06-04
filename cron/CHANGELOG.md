# Changelog

所有版本变更记录。格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)。

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
