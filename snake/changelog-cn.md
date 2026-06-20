# Changelog

[English](./CHANGELOG.md)

所有版本变更记录。格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)。

## [0.0.4] - 2026-06-19

### 新增

- `Validate()` 新增位宽校验：`WorkerIDBits + SequenceBits` 超过 63 时返回错误，避免时间戳字段溢出
- 新增完整单元测试，覆盖率达 97.8%（含并发唯一性、时钟回拨、序列号溢出、位宽边界等场景）
- 新增 CI 工作流（`.github/workflows/ci-snake.yml`），集成 Codecov 覆盖率上报

### 优化

- 移除 `CalculateWorkerID` 中两处死代码（`fnv.Write` 永不返回错误；取模运算保证 workerID 始终在有效范围内）

## [0.0.3] - 2026-06-04

### 依赖升级

- `github.com/zeromicro/go-zero` v1.10.0 → v1.10.2
- `go.opentelemetry.io/otel*` v1.24.0 → v1.40.0（indirect，主动不升 v1.44+ 避免强制要求 go 1.25）
- 同步 `go mod tidy` 清理无用 indirect 依赖
- `go` directive 保持 1.24.0

## [0.0.2] - 2026-03-20

- 升级 Go 版本至 1.24.0
- 升级 go-zero 至 v1.10.0
- 更新依赖版本

## [0.0.1] - 2026-01-26

- 项目初版发布
