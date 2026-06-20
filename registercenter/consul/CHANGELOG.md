# Changelog

[中文](./changelog-cn.md)

All notable changes to this project are recorded here. Format based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [0.1.7] - 2026-06-04

### Dependencies

- `github.com/zeromicro/go-zero` v1.10.0 → v1.10.2
- `google.golang.org/grpc` v1.65.0 → v1.80.0
- `go.opentelemetry.io/otel/trace` v1.24.0 → v1.40.0 (indirect)
- Run `go mod tidy` to clean up unused indirect dependencies
- Pin `github.com/hashicorp/consul/api` to v1.25.1 (v1.33+ requires go 1.25; pinned for go 1.24 compatibility)
- `go` directive kept at 1.24.0

## [0.1.6] - 2026-03-20

- Upgrade Go version to 1.24.0
- Upgrade go-zero to v1.10.0
- Update dependency versions

## [0.1.5] - 2025-12-01

- Optimization release

## [0.1.3] - 2025-12-01

- Optimization release

## [0.1.2] - 2025-12-01

- Optimization release

## [0.1.1] - 2025-12-01

- Optimization release

## [0.1.0] - 2025-12-01

- Initial release
