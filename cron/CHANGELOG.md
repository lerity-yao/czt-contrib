# Changelog

[中文](./changelog-cn.md)

All version change logs. Format based on [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/).

## [0.1.1] - 2026-06-04

### Dependencies

- `github.com/zeromicro/go-zero` v1.10.0 → v1.10.2
- `github.com/redis/go-redis/v9` v9.17.3 → v9.20.0
- Synced `go mod tidy` to clean up unused indirect dependencies
- Intentionally kept `go.opentelemetry.io/otel*` at v1.40.0 (v1.44+ forces go 1.25; not upgraded for compatibility)
- `go` directive remains 1.24.0

## [0.1.0] - 2026-06-04

### Breaking Changes

- **`CronAdd` signature adjustment**: added required parameter `handler HandlerFunc`; completes scheduler registration and handler registration in one step, avoiding repeating the same task type in both `CronAdd` and `Add`
  - Old usage: `CronAdd(spec, pattern, opts...)` + `Add(pattern, handler)`
  - New usage: `CronAdd(spec, pattern, handler, opts...)`
- **Scheduled task timeout now managed by go-zero `RestConf.Timeout`**: before v0.1.0, scheduled tasks did not inject a timeout and fell back to the asynq default hard-coded constant of 30 minutes; from v0.1.0, it is recommended to inject `serverCtx.Config.Timeout` (go-zero `rest.RestConf.Timeout`) into `CronAdd` via `asynq.Timeout(d)` in `workers.go`, so that the `ctx` received by the handler carries a `Deadline`, unified with go-zero HTTP timeout management
  - The cztctl template will be updated in subsequent versions; when upgrading, business teams need to manually adjust `workers.go`. See the migration guide for details

### Migration Guide

**API Signature Migration**

```diff
- serverCtx.CronServer.CronAdd("*/1 * * * *", "GDemoB")
- serverCtx.CronServer.Add("GDemoB", demoA.GDemoBHandler(serverCtx))
+ serverCtx.CronServer.CronAdd("*/1 * * * *", "GDemoB", demoA.GDemoBHandler(serverCtx))
```

**Integrate go-zero Timeout Management**

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

> `rest.RestConf.Timeout` defaults to `3000ms`, which is short for IO-intensive scheduled tasks; it is recommended to explicitly increase it in yaml (e.g., `Timeout: 30000`).

### Added

- `CronAdd` / `Client.Push*` can set an arbitrary timeout per task via the `asynq.Timeout(d)` opt (not limited by the asynq default 30 minutes), e.g., `CronAdd("*/1 * * * *", "X", h, asynq.Timeout(time.Hour))`
- README advanced guide added the "Task Timeout Control" section, detailing the timeout injection mechanism for scheduled tasks and pushed tasks

### Fixed

- `ExponentialRetryDelay` formula corrected to `2^(n+1) - 1`: the `n` passed by asynq is the number of retries already performed (0-indexed); the original formula `2^n - 1` returned 0s on the first retry, causing no backoff on the first retry; after correction the retry sequence is stable at `1s, 3s, 7s, 15s, 31s, ...`

## [0.0.9] - 2026-06-04

### Added

- Added `SetBaseContext(ctx context.Context)` method, supporting injection of a base context; all task handler ctxs will use this as parent

### Changed

- `asynq.Server` is created lazily at `Start()`, supporting dynamic context configuration after construction and before startup

## [0.0.8] - 2026-06-04

### Added

- Added `GetTaskID`, `GetRetryCount`, `GetMaxRetry`, `GetQueueName` wrapper functions; the handler layer can obtain task metadata without directly depending on asynq

### Changed

- README.md fully refactored: full examples organized by go-zero / standalone scripts + Server / Client quadrants
- Advanced guide added task group aggregation and skip retry usage examples
- Added `asynq.SkipRetry` wrapping notes (`pkg/errors` incompatible)
- Metrics section heading levels corrected

## [0.0.7] - 2026-06-04

### Added

- Added `WithGroupAggregator` ServerOption, supporting injection of a task group aggregator to make Group grouping available
- Added `WithRetryDelayFunc` ServerOption, supporting custom retry delay strategy
- Added built-in retry strategy `ExponentialRetryDelay` (`2^n - 1` seconds: 1s, 3s, 7s, 15s, 31s...)
- `CronAdd` automatically uses `realPattern` as TaskID so scheduled tasks are not duplicated across multi-instance deployments

### Fixed

- Removed dead code validation for `GroupGracePeriod` in `ServerConfig.Validate()` (int64 cannot be >0 and <1)

### Changed

- `GroupMaxDelay` default changed from 0 (unlimited) to 300 seconds (5 minutes), preventing tasks from never aggregating when they keep arriving
- Server default retry strategy changed from asynq built-in `n⁴` exponential backoff to `ExponentialRetryDelay` (`2^n - 1` seconds)
- `Add` logs unified to `[CRON] Worker registered: %s`

## [0.0.6] - 2026-05-20

### Changed

- Replaced `log.Printf` with `logx.Errorf` in `metrics.go` to unify with go-zero logging conventions
- Removed `LoggingMiddleware` to avoid duplication with asynq internal error logs
- Removed redundant `logc.Errorf` in `RecoveryMiddleware`; asynq `perform` already has built-in panic recovery logging

## [0.0.5] - 2026-05-15

### Added

- Added full metrics system
  - Server side: `cron_server_consume_total`, `cron_server_consume_duration_ms`,
   `cron_server_consume_bytes`, `cron_server_active_workers`, `cron_server_retry_total`,
   `cron_server_skip_retry_total`, `cron_server_panic_total`, `queue_groups`,
   `tasks_aggregating_total`, `cron_scheduler_trigger_total`, `cron_scheduler_registered`,
   `cron_tasks_enqueued_total`, `cron_queue_size`, `cron_queue_latency_seconds`
  - Client side: `cron_client_push_total`, `cron_client_push_duration_ms`, `cron_client_push_bytes`, `cron_client_cancel_total`
- Added queue whitelist filtering to avoid metric mixing when multiple services share Redis
- Client defaults to no retries (MaxRetry=0); users can override via opts

### Changed

- Metrics namespace changed from `asynq_` to `cron_`
- Server-side metrics added `server` subsystem

## [0.0.4] - 2026-04-10

### Changed

- Upgraded Go version to 1.24.0
- Upgraded go-zero to v1.10.0
- Updated dependencies
