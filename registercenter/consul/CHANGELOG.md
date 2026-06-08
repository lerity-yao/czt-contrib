# Changelog

所有版本变更记录。格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)。

## [0.1.7] - 2026-06-04

### 依赖升级

- `github.com/zeromicro/go-zero` v1.10.0 → v1.10.2
- `google.golang.org/grpc` v1.65.0 → v1.80.0
- `go.opentelemetry.io/otel/trace` v1.24.0 → v1.40.0（indirect）
- 同步 `go mod tidy` 清理无用 indirect 依赖
- 主动锁定 `github.com/hashicorp/consul/api` v1.25.1（v1.33+ 强制要求 go 1.25，为兼容 go 1.24 不升）
- `go` directive 保持 1.24.0

## [0.1.6] - 2026-03-20

- 升级 Go 版本至 1.24.0
- 升级 go-zero 至 v1.10.0
- 更新依赖版本

## [0.1.5] - 2025-12-01

- 优化版本

## [0.1.3] - 2025-12-01

- 优化版本

## [0.1.2] - 2025-12-01

- 优化版本

## [0.1.1] - 2025-12-01

- 优化版本

## [0.1.0] - 2025-12-01

- 项目初版发布
