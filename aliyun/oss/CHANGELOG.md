# Changelog

所有版本变更记录。格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)。

## [0.0.1] - 2026-06-14

### 新增

- 阿里云 OSS Go 客户端，基于 go-zero httpc 封装，自动完成 HMAC-SHA1 V1 签名
- `PutObject` 方法：上传对象，支持 `WithContentType` / `WithMeta` / `WithHeader` 选项
- `GetObject` 方法：下载对象，返回流式 `io.ReadCloser`
- `DeleteObject` 方法：删除对象
- `HeadObject` 方法：获取对象元信息（大小、类型、ETag、自定义元数据）
- `CopyObject` 方法：同 Bucket 内拷贝对象
- `ListObjects` 方法：列举对象，支持 `WithPrefix` / `WithMarker` / `WithMaxKeys` / `WithDelimiter` / `WithEncodingType`
- `Do` 方法：原始签名请求，支持任意 HTTP 方法和自定义 header
- `Parse` 函数：封装 `encoding/xml`，解码 XML 响应并自动关闭 Body
- `WithClient` Option：注入自定义 `*http.Client`（TLS、连接池等）
- `ServiceError` 错误类型：自动解析 OSS XML 错误响应
- 底层集成 go-zero 熔断器，同一 Host 自动共享
