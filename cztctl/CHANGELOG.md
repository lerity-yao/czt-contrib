# Changelog

[中文](./changelog-cn.md)

All notable changes to this project are recorded here. Format based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [1.10.10] - 2026-06-18

### Fixed

- Fixed an issue in `cztctl rpc sdk` where generating `.kong.proto` from a CRLF (Windows) line-ending proto file caused rpc line semicolons to not be removed and annotations to be misaligned
  - Root cause: `strings.Split(data, "\n")` left trailing `\r` at the end of each line, causing `TrimSuffix(";")` to match `\r` instead of `;`, preserving the semicolon and shifting ` {`
  - Fix: Normalize `\r\n` / `\r` to `\n` at read time to eliminate `\r` interference on suffix matching at the source

### Changed

- `rpcMethodRe` regex now uses the `returns` keyword to identify rpc method declarations, which is more semantic than the original `\(` bracket matching (a line with `returns` is a true rpc method)

## [1.10.9] - 2026-06-18

### Fixed

- Fixed an issue in `cztctl rpc sdk` where generating `.kong.proto` with an rpc line using an empty `{}` block caused annotation misalignment
  - The original check `strings.Contains(line, "{")` incorrectly treated `{}` as an already-opened brace, causing the `option (google.api.http)` annotation to float outside the block with an extra closing brace
  - Fixed by first stripping the closing `}` of `{}`, then using `HasSuffix("{")` for detection, covering all three rpc line styles: semicolon / open brace / empty block

## [1.10.8] - 2026-06-16

### Added

- `cztctl rpc sdk` automatically generates `.kong.proto` files (Kong gRPC-gateway HTTP annotations)
  - Automatically generates a `.kong.proto` variant with the same name as the proto in the `_sdk/` directory when `cztctl rpc sdk` is executed
  - Automatically adds `import "google/api/annotations.proto"`
  - Generates `option (google.api.http)` annotations for each rpc method
  - Path rule: `/{ServiceName}/{RpcMethodName}`
  - HTTP method is uniformly POST with `body: "*"`
  - Skips inserting the annotations import if the proto already contains it

## [1.10.7] - 2026-06-04

### Changed

- `cztctl api cron` generator adapted to the new API of [cron v0.1.0](https://github.com/lerity-yao/czt-contrib/blob/main/cron/CHANGELOG.md#010---2026-06-04)
  - `CronAdd` unified signature: pass the handler directly at call time; registration + scheduling are completed internally, no additional `Add` call required
  - Scheduled tasks now default-inject `asynq.Timeout(time.Duration(serverCtx.Config.Timeout) * time.Millisecond)`; timeout is managed by go-zero `RestConf.Timeout`
  - When there are multiple scheduled tasks, a `timeoutOpt` local variable is extracted to avoid repeating the timeout expression
  - `MaxRetry` is still configured per-task via `CronRetry`
  - Non-scheduled tasks (external dispatch consumers) remain unchanged, registering only the handler

### Dependencies

- `github.com/zeromicro/go-zero` v1.10.1 → v1.10.2
- `github.com/zeromicro/go-zero/tools/goctl` stays at v1.10.1 (upstream has not released v1.10.2)

## [1.10.6] - 2026-05-27

### Added

- `cztctl rpc sdk` — automatically generate RPC client code as a standalone Go module and publish it to a Git repository
  - Automatically clone/initialize the SDK repository with HTTPS + Token authentication support
  - Recursively parse proto import dependencies and automatically copy all related proto files
  - Call goctl to generate client code; clean up server-side code to keep only `client/`
  - Automatically run `go mod tidy` to tidy dependencies
  - SemVer version management: auto-increment patch, auto-carry minor when patch reaches 99
  - Automatic Git commit, tagging, and push (branch + tag)
  - Support `--remote` / `--branch` goctl remote template pass-through
  - Support `--multiple` multi-service mode
  - Support `--goproxy` custom proxy
  - Support `--repo-branch` to specify the SDK repository branch (default `main`)
  - goctl version pre-check (not installed is a hard error; version too low is a soft warning)

## [1.10.5] - 2026-04-09

### Fixed

- Fixed an issue in `cztctl api cron` where generating workers.go included a handler argument in `CronAdd` incorrectly
  - Scheduled tasks (`@cron`) now correctly generate two lines: `Add(pattern, handler)` to register the handler, followed by `CronAdd(cronExpr, pattern, opts...)` to register the schedule
  - Pure async tasks (without `@cron`) still generate only `Add(pattern, handler)`
- Fixed a go.sum checksum mismatch for goctl causing `go install` to fail

### Changed

- Upgrade goctl dependency v1.10.0 → v1.10.1
- Upgrade go-zero dependency v1.10.0 → v1.10.1

## [1.10.3] - 2026-04-07

### Added

- `.cron` route names support `-` (hyphen) and `:` (colon) separators, e.g. `sync-order`, `email:send`
- `.rabbitmq` route names support `-` (hyphen) separator, e.g. `payment-refund`, `order.pay-callback`
- Route name separators are semantically isolated by file type: `.cron` only allows `-` `:`, `.rabbitmq` only allows `.` `-`

## [1.10.2] - 2026-03-22

### Added

- `cztctl api swagger` — generate Swagger 2.0 documentation from a .api file
  - Full info block property mapping (title / description / version / host / basePath / schemes, etc.)
  - Support @server annotations (tags / summary / prefix / group / deprecated / operationId / authType)
  - Support automatic `validate` tag parameter constraint annotations
  - Support multi-line field header comment parsing
  - Support `useDefinitions` mode (`$ref` references)
  - Support `wrapCodeMsg` response wrapping
  - Support both JSON and YAML output formats
- `cztctl api cron` — generate a distributed cron job service from a .cron file
  - Based on the [czt-contrib/cron](https://github.com/lerity-yao/czt-contrib/cron) framework
  - Support internal scheduled tasks (`@cron` expression + `@cronRetry` retries)
  - Support externally triggered tasks (no `@cron`, triggered via asynq.Client.Enqueue)
  - Generate complete directory structure: etc / config / handler / logic / svc / types / worker / main
  - Support `@doc` in both string and KV forms
  - Support `@server` grouping (group / tags / summary / middleware)
  - Support `--remote` / `--branch` remote templates
  - Support `--style` file naming style (gozero / go_zero / goZero)
- `cztctl api rabbitmq` — generate a RabbitMQ consumer service from a .rabbitmq file
  - Based on the [czt-contrib/mq/rabbitmq](https://github.com/lerity-yao/czt-contrib/mq/rabbitmq) framework
  - Support dot-separated queue names (e.g. `order.created`, `payment.refund.success`)
  - Support optional message parameter types
  - Generate complete directory structure: etc / config / handler / logic / svc / types / listener / main
  - Support `--remote` / `--branch` remote templates
  - Support `--style` file naming style
- `cztctl env` — environment variable management
  - View current environment variables (CZTCTL_OS / CZTCTL_ARCH / CZTCTL_HOME / CZTCTL_CACHE / CZTCTL_VERSION)
  - Edit environment variables (`cztctl env -w KEY=VALUE`)
  - `CZTCTL_EXPERIMENTAL` toggle: `off` uses ANTLR4 parser, `on` uses the handwritten recursive-descent parser
- DSL syntax parsing
  - `.cron` and `.rabbitmq` share the base syntax: syntax / info / import / type / @server
  - ANTLR4 parser (default) + handwritten recursive-descent parser (experimental)
  - Full type system support: primitive types, slices, maps, pointers, nested structs, struct tags
  - Support cross-file type references via import
- Versioning rule: `v<go-zero-major-version>.<micro-version>` (currently based on go-zero v1.10.0)
