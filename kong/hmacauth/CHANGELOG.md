# Changelog

[中文](./changelog-cn.md)

All version change records. Format based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [0.0.1] - 2026-06-18

### Added

- Kong HMAC Auth Go client, built on go-zero httpc, with automatic HMAC signing, following the official [Kong HMAC Auth plugin](https://developer.konghq.com/plugins/hmac-auth/) specification
- `Do` method: structured requests with `path` / `form` / `json` / `header` tag auto-mapping
- `DoRaw` method: raw byte requests for file uploads, XML, plain text, and custom body
- `WithClient` option: inject a custom `*http.Client` (TLS, connection pool, etc.)
- `Parse` function: wraps go-zero `httpc.Parse`, accepts `(resp, err)` from `Do`/`DoRaw`, auto-parses JSON response and closes Body
- Support for 5 HMAC algorithms: `hmac-sha1`, `hmac-sha224` (Kong 3.14+), `hmac-sha256`, `hmac-sha384`, `hmac-sha512`
- `@request-target` pseudo-header support, value is `method /path?query` (method in lowercase)
- `Conf.Headers` allows custom header list for signing, defaults to `["date", "@request-target"]`
- `Digest` header auto-computation (`SHA-256=base64(sha256(body))`), supports empty body zero-length digest, form / multipart skipped
- `Date` header auto-injection (GMT format), used for Kong clock skew validation
- `User-Agent` header auto-injection (`Go-Kong-HmacAuth-Client`), can be overridden by the caller
- `Host` header injection and signing value uniformly use `r.Host`, ensuring consistency between signed and actually sent values
- `Conf.Algorithm` and `Conf.Headers` support go-zero `optional` / `default` tags, can be omitted in YAML configuration
- Built-in go-zero circuit breaker, tracing, and metrics integration, auto-shared per Host
