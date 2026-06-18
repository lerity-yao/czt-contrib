# Changelog

所有版本变更记录。格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)。

## [1.10.9] - 2026-06-18

### 修复

- 修复 `cztctl rpc sdk` 生成 `.kong.proto` 时，rpc 行使用 `{}` 空块写法导致注解错位的问题
  - 原判断 `strings.Contains(line, "{")` 将 `{}` 空块误判为已有开括号，导致 `option (google.api.http)` 注解游离到块外并多出闭合括号
  - 改为先剥除 `{}` 的闭合括号，再用 `HasSuffix("{")` 判断，覆盖分号 / 开括号 / 空块三种 rpc 行风格

## [1.10.8] - 2026-06-16

### 新增

- `cztctl rpc sdk` 自动生成 `.kong.proto` 文件（Kong gRPC-gateway HTTP 注解）
  - 执行 `cztctl rpc sdk` 时自动在 `_sdk/` 目录中生成与 proto 同名的 `.kong.proto` 变体
  - 自动添加 `import "google/api/annotations.proto"`
  - 为每个 rpc 方法生成 `option (google.api.http)` 注解
  - 路径规则：`/{ServiceName}/{RpcMethodName}`
  - HTTP 方法统一 POST + `body: "*"`
  - 若 proto 已包含 annotations import 则跳过插入

## [1.10.7] - 2026-06-04

### 变更

- `cztctl api cron` 生成器适配 [cron v0.1.0](https://github.com/lerity-yao/czt-contrib/blob/main/cron/CHANGELOG.md#010---2026-06-04) 新 API
  - `CronAdd` 一体化签名：调用时直接传入 handler，内部完成注册 + 调度，不再需要额外 `Add`
  - 定时任务默认注入 `asynq.Timeout(time.Duration(serverCtx.Config.Timeout) * time.Millisecond)`，超时由 go-zero `RestConf.Timeout` 接管
  - 多个定时任务时抽取 `timeoutOpt` 局部变量，避免重复书写超时表达式
  - `MaxRetry` 仍按 per-task `CronRetry` 配置
  - 非定时任务（外部投递消费端）保持仅注册 handler，不变

### 依赖升级

- `github.com/zeromicro/go-zero` v1.10.1 → v1.10.2
- `github.com/zeromicro/go-zero/tools/goctl` 保持 v1.10.1（上游未跟版发 v1.10.2）

## [1.10.6] - 2026-05-27

### 新增

- `cztctl rpc sdk` — 将 RPC 客户端代码自动生成为独立 Go 模块并发布到 Git 仓库
  - 自动克隆/初始化 SDK 仓库，支持 HTTPS + Token 认证
  - 递归解析 proto import 依赖，自动复制所有相关 proto 文件
  - 调用 goctl 生成客户端代码，清理服务端代码仅保留 client/
  - 自动执行 `go mod tidy` 整理依赖
  - SemVer 版本号管理：自动递增 patch，patch 到 99 自动进位 minor
  - Git 自动提交、打标签、推送（分支 + 标签）
  - 支持 `--remote` / `--branch` goctl 远程模板透传
  - 支持 `--multiple` 多服务模式
  - 支持 `--goproxy` 自定义代理
  - 支持 `--repo-branch` 指定 SDK 仓库分支（默认 main）
  - goctl 版本前置检查（未安装为硬错误，版本低为软警告）

## [1.10.5] - 2026-04-09

### 修复

- 修复 `cztctl api cron` 生成 workers.go 时，`CronAdd` 错误包含 handler 参数的问题
  - 定时任务（`@cron`）现在正确生成两行：先 `Add(pattern, handler)` 注册处理函数，再 `CronAdd(cronExpr, pattern, opts...)` 注册定时调度
  - 纯异步任务（无 `@cron`）仍只生成 `Add(pattern, handler)`
- 修复 go.sum 中 goctl 校验和不匹配导致 `go install` 失败的问题

### 变更

- 升级 goctl 依赖 v1.10.0 → v1.10.1
- 升级 go-zero 依赖 v1.10.0 → v1.10.1

## [1.10.3] - 2026-04-07

### 新增

- .cron 路由名支持 `-`（横杠）和 `:`（冒号）分隔符，如 `sync-order`、`email:send`
- .rabbitmq 路由名支持 `-`（横杠）分隔符，如 `payment-refund`、`order.pay-callback`
- 路由名分隔符按文件类型语义隔离：.cron 仅允许 `-` `:`，.rabbitmq 仅允许 `.` `-`

## [1.10.2] - 2026-03-22

### 新增

- `cztctl api swagger` — 从 .api 文件生成 Swagger 2.0 文档
  - 支持 info 块全量属性映射（title / description / version / host / basePath / schemes 等）
  - 支持 @server 注解（tags / summary / prefix / group / deprecated / operationId / authType）
  - 支持 `validate` tag 自动生成参数约束注释
  - 支持字段头部多行注释解析
  - 支持 `useDefinitions` 模式（`$ref` 引用）
  - 支持 `wrapCodeMsg` 响应包装
  - 支持 JSON / YAML 双格式输出
- `cztctl api cron` — 从 .cron 文件生成分布式定时任务服务
  - 基于 [czt-contrib/cron](https://github.com/lerity-yao/czt-contrib/cron) 框架
  - 支持内部定时任务（`@cron` 表达式 + `@cronRetry` 重试）
  - 支持外部触发任务（无 `@cron`，通过 asynq.Client.Enqueue 触发）
  - 生成完整目录结构：etc / config / handler / logic / svc / types / worker / main
  - 支持 `@doc` 字符串与 KV 两种写法
  - 支持 `@server` 分组（group / tags / summary / middleware）
  - 支持 `--remote` / `--branch` 远程模板
  - 支持 `--style` 文件命名风格（gozero / go_zero / goZero）
- `cztctl api rabbitmq` — 从 .rabbitmq 文件生成 RabbitMQ 消费者服务
  - 基于 [czt-contrib/mq/rabbitmq](https://github.com/lerity-yao/czt-contrib/mq/rabbitmq) 框架
  - 支持点分隔队列名称（如 `order.created`、`payment.refund.success`）
  - 支持可选消息参数类型
  - 生成完整目录结构：etc / config / handler / logic / svc / types / listener / main
  - 支持 `--remote` / `--branch` 远程模板
  - 支持 `--style` 文件命名风格
- `cztctl env` — 环境变量管理
  - 查看当前环境变量（CZTCTL_OS / CZTCTL_ARCH / CZTCTL_HOME / CZTCTL_CACHE / CZTCTL_VERSION）
  - 编辑环境变量（`cztctl env -w KEY=VALUE`）
  - `CZTCTL_EXPERIMENTAL` 开关：`off` 使用 ANTLR4 解析器，`on` 使用手写递归下降解析器
- DSL 语法解析
  - .cron 和 .rabbitmq 共享基础语法：syntax / info / import / type / @server
  - ANTLR4 解析器（默认）+ 手写递归下降解析器（实验性）
  - 支持完整类型系统：基本类型、切片、map、指针、嵌套结构体、struct tag
  - 支持 import 跨文件类型引用
- 版本号规则：`v<go-zero主版本>.<微版本>`（当前基于 go-zero v1.10.0）
