# consul

[English](./README.md)

[![codecov](https://codecov.io/gh/lerity-yao/czt-contrib/branch/main/graph/badge.svg?flag=registercenter-consul)](https://codecov.io/gh/lerity-yao/czt-contrib)

基于 [Consul](https://developer.hashicorp.com/consul) 和 [go-zero](https://github.com/zeromicro/go-zero) 的服务注册与发现模块，支持自动注册、健康检查（TTL / HTTP / gRPC）、自动恢复和 gRPC 服务发现解析器。

## 特性

- 📋 **自动注册与注销** — 服务启动自动注册，进程退出通过 `proc.AddShutdownListener` 自动注销
- 💓 **多种健康检查** — 支持 TTL、HTTP、gRPC 三种健康检查机制
- 🔄 **自动恢复** — 健康检查失败时自动重试注册，采用指数退避策略
- 🔍 **gRPC 服务发现** — 内置 `consul://` scheme 解析器，`init()` 自动注册，支持阻塞查询和标签过滤
- 🐳 **容器环境适配** — 自动检测 `POD_IP` 环境变量（Kubernetes），回退到内部 IP
- 🔧 **可扩展监控** — 通过 `WithMonitorFuncs` 注入自定义监控函数

## 安装

```bash
go get github.com/lerity-yao/czt-contrib/registercenter/consul@latest
```

## 配置参数

### Conf

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|:----:|--------|------|
| `Host` | string | 是 | - | Consul 服务器地址，格式 `host:port`，例如 `127.0.0.1:8500` |
| `Key` | string | 是 | - | 服务名称，例如 `user-service` |
| `Scheme` | string | 否 | `http` | 连接协议，可选 `http` / `https` |
| `Token` | string | 否 | `""` | Consul ACL 访问令牌 |
| `Tag` | []string | 否 | `[]` | 服务标签列表 |
| `Meta` | map[string]string | 否 | `nil` | 服务元数据 |
| `TTL` | int | 否 | `20` | 健康检查间隔（秒）。TTL 模式下为心跳间隔；HTTP / gRPC 模式下为 Consul 服务端发起检查的间隔 |
| `ExpiredTTL` | int | 否 | `3` | 服务注销系数。实际注销时间为 `TTL * ExpiredTTL` 秒 |
| `CheckTimeout` | int | 否 | `3` | 健康检查超时时间（秒）。仅 HTTP / gRPC 模式生效（TTL 模式不使用） |
| `CheckType` | string | 否 | `ttl` | 健康检查类型，可选 `ttl` / `http` / `grpc` |
| `CheckHttp` | [CheckHttpConf](#checkhttpconf) | 否 | - | HTTP 健康检查配置，`CheckType` 为 `http` 时生效 |
| `CheckGrpc` | [CheckGrpcConf](#checkgrpcconf) | 否 | - | gRPC 健康检查配置，`CheckType` 为 `grpc` 时生效 |

> 调用 `NewService` 时会自动执行 `Conf.Validate()` 校验上述字段。

### CheckHttpConf

| 参数名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `Method` | string | `GET` | HTTP 方法，可选 `GET` / `POST` |
| `Path` | string | `/healthz` | 健康检查路径 |
| `Host` | string | `0.0.0.0` | 健康检查主机地址 |
| `Port` | int | `6060` | 健康检查端口 |
| `Scheme` | string | `http` | HTTP 协议，可选 `http` / `https` |

### CheckGrpcConf

| 参数名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `TLSServerName` | string | `""` | TLS 服务器名称（可选），用于 TLS 连接验证 |
| `TLSSkipVerify` | bool | `true` | 是否跳过 TLS 验证 |
| `GRPCUseTLS` | bool | `false` | 是否使用 TLS 连接 |

## API 参考

### 构造函数

| 函数 | 签名 | 说明 |
|------|------|------|
| `MustNewService` | `func MustNewService(listenOn string, c Conf, opts ...ServiceOption) Client` | 创建服务实例，校验失败 panic |
| `NewService` | `func NewService(listenOn string, c Conf, opts ...ServiceOption) (Client, error)` | 创建服务实例，校验失败返回 error |

> `listenOn` 为服务监听地址，例如 `:8080` 或 `0.0.0.0:8080`。模块会自动解析为实际可访问的 IP:Port。

### ServiceOption

| Option | 参数 | 说明 |
|--------|------|------|
| `WithMonitorFuncs` | `funcs ...MonitorFunc` | 注入自定义监控函数。不传时根据 `CheckType` 自动选择默认监控函数 |

### Client 接口方法

| 方法 | 签名 | 说明 |
|------|------|------|
| `RegisterService` | `RegisterService() error` | 注册服务并启动健康监控，自动注册优雅关闭回调 |
| `DeregisterService` | `DeregisterService() error` | 注销服务并停止所有监控协程 |
| `GetServiceID` | `GetServiceID() string` | 获取服务 ID，格式为 `Key-Host-Port` |
| `GetRegistration` | `GetRegistration() *api.AgentServiceRegistration` | 获取服务注册信息 |
| `GetServiceClient` | `GetServiceClient() *api.Client` | 获取 Consul API 客户端实例 |

### 监控函数

| 函数 | 签名 | 说明 |
|------|------|------|
| `TTLCheckMonitorFunc` | `func TTLCheckMonitorFunc() MonitorFunc` | TTL 健康检查默认监控函数，定期调用 `UpdateTTL` 更新心跳 |
| `HttpCheckMonitorFunc` | `func HttpCheckMonitorFunc() MonitorFunc` | HTTP / gRPC 健康检查默认监控函数，定期查询服务健康状态 |
| `TTLMonitorLogic` | `func TTLMonitorLogic(cc *CommonClient, state *MonitorState) error` | TTL 监控逻辑，包含自动重试注册 |
| `HttpMonitorLogic` | `func HttpMonitorLogic(cc *CommonClient, state *MonitorState) error` | HTTP / gRPC 监控逻辑，包含自动重试注册 |

> **默认监控函数选择规则**：`CheckType` 为 `ttl` 时使用 `TTLCheckMonitorFunc()`；为 `http` 或 `grpc` 时均使用 `HttpCheckMonitorFunc()`。

### 公开类型

| 类型 | 定义 | 说明 |
|------|------|------|
| `MonitorFunc` | `func(cc *CommonClient, stopChan <-chan struct{})` | 监控函数签名，接收 `CommonClient` 和停止通道 |
| `ServiceOption` | `func(*CommonClient)` | 服务选项函数签名 |
| `MonitorState` | `struct{...}` | 监控状态，包含重试计数、退避时间、Ticker 等，提供 `Close()` 方法 |

### 常量

| 常量 | 值 | 说明 |
|------|------|------|
| `CheckTypeTTL` | `"ttl"` | TTL 健康检查类型 |
| `CheckTypeHttp` | `"http"` | HTTP 健康检查类型 |
| `CheckTypeGrpc` | `"grpc"` | gRPC 健康检查类型 |

## 进阶指南

### 健康检查机制

支持三种健康检查类型，由 `Conf.CheckType` 决定：

#### TTL 检查（`ttl`）

- **机制**：服务定期向 Consul 发送 `UpdateTTL` 心跳以维持健康状态
- **心跳频率**：`TTL - 1` 秒（最小 1 秒）
- **适用场景**：需要应用自定义健康逻辑的场景
- Consul 在 `TTL * ExpiredTTL` 秒内未收到心跳，将自动注销服务

#### HTTP 检查（`http`）

- **机制**：Consul 服务端按 `TTL` 秒间隔主动向服务的健康检查端点发起 HTTP 请求
- **超时**：由 `CheckTimeout` 控制
- **适用场景**：有 Web 接口的服务
- go-zero 开启健康检查后，默认提供 `host:6060/healthz` 端点

```go
conf := consul.Conf{
    Host:      "127.0.0.1:8500",
    Key:       "user-service",
    CheckType: consul.CheckTypeHttp,
    TTL:       20,
    CheckTimeout: 3,
    CheckHttp: consul.CheckHttpConf{
        Method: "GET",
        Path:   "/healthz",
        Host:   "0.0.0.0",
        Port:   6060,
        Scheme: "http",
    },
}
```

#### gRPC 检查（`grpc`）

- **机制**：Consul 服务端按 `TTL` 秒间隔主动向服务发起 gRPC 健康检查请求
- **超时**：由 `CheckTimeout` 控制
- **适用场景**：gRPC 服务
- go-zero 开启 rpc 服务后，默认提供 `grpc.health.v1.Health/Check` 端点

```go
conf := consul.Conf{
    Host:      "127.0.0.1:8500",
    Key:       "user-service",
    CheckType: consul.CheckTypeGrpc,
    TTL:       20,
    CheckTimeout: 5,
    CheckGrpc: consul.CheckGrpcConf{
        TLSServerName: "example.com", // 可选
        TLSSkipVerify: true,
        GRPCUseTLS:    false,
    },
}
```

> 使用 gRPC 检查时，服务需实现标准的健康检查接口（`grpc.health.v1.Health`）。

### 自动恢复机制

当健康检查失败时（TTL 更新失败或 HTTP / gRPC 健康状态不为 `passing`），监控协程会自动尝试重新注册服务：

| 参数 | 值 | 说明 |
|------|------|------|
| 最大重试次数 | 5 | 超过后重置计数器继续尝试 |
| 初始退避时间 | 1 秒 | 首次重试的等待时间 |
| 最大退避时间 | 30 秒 | 退避时间上限 |
| 退避策略 | 指数退避 | `backoff * 2`，超过上限则取上限 |

重试流程：
1. 健康检查失败
2. 检查服务当前健康状态是否为 `passing`
3. 若不是且未达最大重试次数，调用 `registerServiceWithPassingHealth()` 重新注册
4. 重试失败则增加退避时间，重试成功则重置计数器并恢复原始心跳频率

### 容器环境适配

模块通过 `figureOutListenOn` 自动解析服务可访问地址：

1. 检查 `POD_IP` 环境变量（Kubernetes 容器环境注入）
2. 使用 go-zero `netx.InternalIp()` 获取系统内部 IP
3. 回退到配置的监听地址

> 当 `listenOn` 的 host 为 `0.0.0.0` 时触发地址解析。非 `0.0.0.0` 的地址保持不变。

### 服务发现 URL 参数

通过 `init()` 自动注册 `consul://` scheme 的 gRPC 解析器。URL 格式：

```
consul://[user:passwd]@host/service?param=value
```

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `healthy` | bool | `false` | 是否只查询健康服务 |
| `tag` | string | `""` | 服务标签过滤 |
| `wait` | duration | - | Consul 阻塞查询等待时间 |
| `timeout` | duration | - | 查询超时时间 |
| `max-backoff` | duration | `1s` | 拉取失败时的最大退避时间 |
| `near` | string | `_agent` | 按距离排序，就近访问 |
| `limit` | int | `0` | 限制返回服务数量（0 = 不限制） |
| `insecure` | bool | `false` | 是否跳过 TLS 验证 |
| `token` | string | `""` | Consul ACL 访问令牌 |
| `dc` | string | `""` | 数据中心 |
| `allow-stale` | bool | `false` | 是否允许返回过期数据 |
| `require-consistent` | bool | `false` | 是否要求一致性强一致读 |

### 优雅关闭

`RegisterService()` 内部通过 `proc.AddShutdownListener` 注册了关闭回调，程序退出时自动执行：

1. 停止所有监控协程（关闭 stop channel）
2. 调用 `ServiceDeregister` 注销服务
3. 记录注销结果日志

> 在 go-zero 环境中无需手动注销；在非 go-zero 环境中，可使用 `defer service.DeregisterService()` 确保注销。

## 完整示例

### 在 go-zero 中使用

```go
// internal/config/config.go
type Config struct {
    rest.RestConf
    Consul consul.Conf
}
```

```yaml
# etc/config.yaml
Name: user-api
Host: 0.0.0.0
Port: 8888

Consul:
  Host: 127.0.0.1:8500
  Key: user-service
  CheckType: ttl
  TTL: 20
  Tag:
    - v1
    - grpc
```

```go
// internal/svc/servicecontext.go
type ServiceContext struct {
    Config    config.Config
    ConsulSrv consul.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
    consulSrv := consul.MustNewService(
        fmt.Sprintf("%s:%d", c.Host, c.Port),
        c.Consul,
    )

    if err := consulSrv.RegisterService(); err != nil {
        logx.Must(err)
    }

    return &ServiceContext{
        Config:    c,
        ConsulSrv: consulSrv,
    }
}
```

### 独立使用

```go
package main

import (
    "fmt"

    "github.com/lerity-yao/czt-contrib/registercenter/consul"
)

func main() {
    conf := consul.Conf{
        Host:      "127.0.0.1:8500",
        Key:       "user-service",
        CheckType: consul.CheckTypeTTL,
        TTL:       20,
        Tag:       []string{"v1", "grpc"},
    }

    service := consul.MustNewService(":8080", conf)

    if err := service.RegisterService(); err != nil {
        panic(err)
    }
    // 非 go-zero 环境需手动注销
    defer service.DeregisterService()

    fmt.Println("service registered:", service.GetServiceID())
    // 启动你的服务...
}
```

### 自定义监控函数

```go
conf := consul.Conf{
    Host:      "127.0.0.1:8500",
    Key:       "user-service",
    CheckType: consul.CheckTypeTTL,
    TTL:       20,
}

// 自定义监控函数
func customMonitorFunc() consul.MonitorFunc {
    return func(cc *consul.CommonClient, stopCh <-chan struct{}) {
        // 你的自定义监控逻辑
    }
}

service, _ := consul.NewService(":8080", conf,
    consul.WithMonitorFuncs(
        consul.TTLCheckMonitorFunc(),   // 保留默认 TTL 监控
        customMonitorFunc(),             // 追加自定义监控
    ),
)

service.RegisterService()
```

### 服务发现（gRPC 客户端）

```go
import (
    "google.golang.org/grpc"
    _ "github.com/lerity-yao/czt-contrib/registercenter/consul" // 自动注册解析器
)

func main() {
    // 使用 consul URL 创建 gRPC 连接
    conn, err := grpc.Dial(
        "consul://127.0.0.1:8500/user-service?healthy=true&tag=v1",
        grpc.WithInsecure(),
        grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
    )
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    // 创建 gRPC 客户端并使用
    // ...
}
```

## 更新日志

查看 [CHANGELOG.md](./CHANGELOG.md)
