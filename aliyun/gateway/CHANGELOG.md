# Changelog

所有版本变更记录。格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)。

## [0.0.1] - 2026-06-13

### 新增

- 阿里云 API 网关 Go 客户端，基于 go-zero httpc 封装，自动完成 HMAC-SHA256 v1 签名
- `Do` 方法：结构化请求，支持 `path` / `form` / `json` / `header` tag 自动映射
- `DoRaw` 方法：原始字节请求，支持文件上传、XML、纯文本等自定义 body
- `WithClient` Option：注入自定义 `*http.Client`（TLS、连接池等）
- `Parse` 函数：封装 go-zero `httpc.Parse`，接收 `Do`/`DoRaw` 返回的 `(resp, err)`，自动解析 JSON 响应并关闭 Body
- `NewMultipart()` / `MultipartBuilder`：链式构造 multipart/form-data 请求体
- `Conf.Validate()` 自动去除 Host 尾部多余 `/`
- 底层集成 go-zero 熔断器，同一 Host 自动共享
