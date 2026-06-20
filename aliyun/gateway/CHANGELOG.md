# Changelog

[中文](./changelog-cn.md)

All version change records. Format based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [0.0.3] - 2026-06-14

### Fixed

- `MultipartBuilder.File` ignoring `CreateFormFile` error could cause nil panic; changed to error accumulation mode
- `MultipartBuilder.Build()` signature changed to `(contentType string, body []byte, err error)`, uniformly returning build errors

### Improved

- Replaced `uuid.New().String()` with `uuid.NewString()` to reduce intermediate UUID struct allocation

## [0.0.2] - 2026-06-14

### Improved

- `sortedQuery` now uses `strings.IndexByte` for zero-allocation iteration, eliminating `strings.Split` `[]string` allocation
- `sortedQuery` eliminates redundant `result` slice, uses `strings.Builder` for direct output; map/slice pre-allocated with capacity
- `signOption` leverages `GetBody` for form/multipart body to skip full buffering, avoiding unnecessary memory copies during large file uploads
- `[]byte(AppSecret)` is pre-computed once outside the closure at `NewClient` time, no longer converted on every signature
- `strings.Join(signHeaders)` promoted to package-level `signHeadersValue`, computed only once

### Fixed

- `signRequest` now shares a single `time.Now()` call for Date and Timestamp, eliminating the risk of inconsistency across millisecond boundaries

## [0.0.1] - 2026-06-13

### Added

- Alibaba Cloud API Gateway Go client, built on go-zero httpc, with automatic HMAC-SHA256 v1 signing
- `Do` method: structured requests with `path` / `form` / `json` / `header` tag auto-mapping
- `DoRaw` method: raw byte requests for file uploads, XML, plain text, and custom body
- `WithClient` option: inject a custom `*http.Client` (TLS, connection pool, etc.)
- `Parse` function: wraps go-zero `httpc.Parse`, accepts `(resp, err)` from `Do`/`DoRaw`, auto-parses JSON response and closes Body
- `NewMultipart()` / `MultipartBuilder`: chain-build multipart/form-data request body
- `Conf.Validate()` automatically strips trailing `/` from Host
- Built-in go-zero circuit breaker integration, auto-shared per Host
