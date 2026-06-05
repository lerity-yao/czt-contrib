# Changelog

所有版本变更记录。格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)。

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
