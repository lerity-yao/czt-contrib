# Changelog

所有版本变更记录。格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)。

## [0.0.1] - 2026-06-18

### 新增

- Kong HMAC Auth Go 客户端，基于 go-zero httpc 封装，自动完成 HMAC 签名，遵循 [Kong HMAC Auth 插件](https://developer.konghq.com/plugins/hmac-auth/) 官方规范
- `Do` 方法：结构化请求，支持 `path` / `form` / `json` / `header` tag 自动映射
- `DoRaw` 方法：原始字节请求，支持文件上传、XML、纯文本等自定义 body
- `WithClient` Option：注入自定义 `*http.Client`（TLS、连接池等）
- `Parse` 函数：封装 go-zero `httpc.Parse`，接收 `Do`/`DoRaw` 返回的 `(resp, err)`，自动解析 JSON 响应并关闭 Body
- 支持 5 种 HMAC 算法：`hmac-sha1`、`hmac-sha224`（Kong 3.14+）、`hmac-sha256`、`hmac-sha384`、`hmac-sha512`
- `@request-target` 伪头支持，值为 `method /path?query`（method 小写）
- `Conf.Headers` 可自定义参与签名的 header 列表，默认 `["date", "@request-target"]`
- `Digest` 头自动计算（`SHA-256=base64(sha256(body))`），支持空 body 零长度摘要，form / multipart 跳过
- `Date` 头自动注入（GMT 格式），用于 Kong clock skew 校验
- `User-Agent` 头自动注入（`Go-Kong-HmacAuth-Client`），可被调用方覆盖
- `Host` 头注入与签名取值统一使用 `r.Host`，确保签名值与实际发送一致
- `Conf.Algorithm` 和 `Conf.Headers` 支持 go-zero `optional` / `default` tag，YAML 配置可省略
- 底层集成 go-zero 熔断器、tracing、metrics，同一 Host 自动共享
