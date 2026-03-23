# Changelog

## v0.0.5

### 新增
- 新增完整监控指标体系
  - Server 端：`cron_server_consume_total`、`cron_server_consume_duration_ms`、
   `cron_server_consume_bytes`、`cron_server_active_workers`、`cron_server_retry_total`、
   `cron_server_skip_retry_total`、`cron_server_panic_total`、`queue_groups`、
   `tasks_aggregating_total`、`cron_scheduler_trigger_total`、`cron_scheduler_registered`、
   `cron_tasks_enqueued_total`、`cron_queue_size`、`cron_queue_latency_seconds`
  - Client 端：`cron_client_push_total`、`cron_client_push_duration_ms`、`cron_client_push_bytes`、`cron_client_cancel_total`、
- 新增队列白名单过滤，解决多服务共用 Redis 时指标混杂问题
- Client 默认不重试（MaxRetry=0），用户可通过 opts 覆盖

### 变更
- 监控指标命名空间从 `asynq_` 改为 `cron_`
- Server 端指标添加 `server` subsystem

## v0.0.4

- 升级 Go 版本至 1.24.0
- 升级 go-zero 至 v1.10.0
- 更新依赖版本
