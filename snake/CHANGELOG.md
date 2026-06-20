# Changelog

[中文](./changelog-cn.md)

All version change logs. Format based on [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/).

## [0.0.4] - 2026-06-19

### Added

- `Validate()` now checks bit width: returns an error when `WorkerIDBits + SequenceBits` exceeds 63 to avoid timestamp field overflow
- Added full unit tests with 97.8% coverage (concurrent uniqueness, clock backward, sequence overflow, bit-width boundaries, etc.)
- Added CI workflow (`.github/workflows/ci-snake.yml`) integrating Codecov coverage reporting

### Improved

- Removed two pieces of dead code in `CalculateWorkerID` (`fnv.Write` never returns an error; the modulo operation guarantees workerID stays within the valid range)

## [0.0.3] - 2026-06-04

### Dependencies

- `github.com/zeromicro/go-zero` v1.10.0 → v1.10.2
- `go.opentelemetry.io/otel*` v1.24.0 → v1.40.0 (indirect; intentionally not upgraded to v1.44+ to avoid forcing go 1.25)
- Synced `go mod tidy` to clean up unused indirect dependencies
- `go` directive remains 1.24.0

## [0.0.2] - 2026-03-20

- Upgraded Go version to 1.24.0
- Upgraded go-zero to v1.10.0
- Updated dependencies

## [0.0.1] - 2026-01-26

- Initial project release
