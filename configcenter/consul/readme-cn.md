# configcenter/consul

[English](./README.md)

基于 [Consul KV](https://developer.hashicorp.com/consul/api-docs/kv) 的 go-zero 配置中心订阅模块，自动监听 KV 变更并通知业务层，实现了 go-zero `configcenter.Subscriber` 接口。

## 特性

- 🔍 **自动监听变更** — 基于 Consul blocking query 长轮询，KV 变更实时触发回调
- 📄 **多格式支持** — 支持 YAML、JSON、HCL、XML 四种配置格式，统一输出 JSON
- 🔌 **无缝集成 go-zero** — 实现 `configcenter.Subscriber` 接口，配合 `configurator.MustNewConfigCenter` 开箱即用
- 🔒 **TLS 与 ACL** — 支持 Consul Token 鉴权和 TLS 加密连接

## 安装

```bash
go get github.com/lerity-yao/czt-contrib/configcenter/consul@v0.1.2
```

## 配置参数

### Conf

| 参数名 | 类型 | 必填 | 默认值 | 说明 |
|--------|------|------|--------|------|
| `Host` | string | 否 | - | Consul 地址，格式 `host:port`，如 `127.0.0.1:8500` |
| `Scheme` | string | 否 | `http` | Consul 地址协议，`http` 或 `https` |
| `PathPrefix` | string | 否 | - | Consul API 路径前缀 |
| `Datacenter` | string | 否 | - | Consul 数据中心名称 |
| `Token` | string | 否 | - | Consul ACL Token |
| `TLSConfig` | `api.TLSConfig` | 否 | - | Consul TLS 连接配置 |
| `Key` | string | 否 | - | Consul KV 路径，即配置在 KV 中的 key，如 `DemoA.api` |
| `Type` | string | 否 | `yaml` | 配置值格式，可选 `yaml`、`hcl`、`json`、`xml` |

> `ConsulConf` 是 `Conf` 的类型别名（`type ConsulConf Conf`），两者等价，推荐使用 `ConsulConf`。

## API 参考

### 构造函数

| 函数 | 签名 | 说明 |
|------|------|------|
| `MustNewConsulSubscriber` | `func MustNewConsulSubscriber(conf ConsulConf) *ConsulSubscriber` | 创建 Subscriber，失败 panic |
| `NewConsulSubscriber` | `func NewConsulSubscriber(conf ConsulConf) (*ConsulSubscriber, error)` | 创建 Subscriber，失败返回 error |

### ConsulSubscriber 方法

| 方法 | 签名 | 说明 |
|------|------|------|
| `Value` | `func (s *ConsulSubscriber) Value() (string, error)` | 从 Consul KV 读取当前值，解析为 JSON 字符串返回 |
| `AddListener` | `func (s *ConsulSubscriber) AddListener(listener func()) error` | 注册变更回调，KV 发生变化时自动触发 |
| `Stop` | `func (s *ConsulSubscriber) Stop()` | 停止后台 watch 协程 |

> `Value()` 和 `AddListener()` 共同实现了 go-zero `configcenter.Subscriber` 接口。

### ConsulSubscriber 导出字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `Path` | string | Consul KV 路径，创建时由 `Conf.Key` 赋值 |
| `Type` | string | 配置格式，创建时由 `Conf.Type` 赋值 |

## 进阶指南

### 监听机制

`ConsulSubscriber` 创建时自动启动后台 watch 协程，基于 Consul KV 的 **blocking query** 机制实现变更监听：

1. **首次请求**：发起 KV Get 请求，记录 `X-Consul-Index`（即 `LastIndex`）
2. **长轮询**：后续请求携带 `WaitIndex=LastIndex`，Consul 在值变更前会阻塞（默认无超时限制）
3. **变更检测**：当 KV 值被修改，Consul 返回新值和更大的 `LastIndex`
4. **通知回调**：`LastIndex` 增大时触发所有已注册的 listener
5. **错误重试**：请求失败时等待 1 秒后重试

```
┌─────────────┐     KV Get (WaitIndex=N)     ┌─────────────┐
│  Subscriber  │ ──────────────────────────► │   Consul    │
│  (watch)     │ ◄────────────────────────── │   KV Store  │
└──────┬──────┘   返回新值 + LastIndex=N+1   └─────────────┘
       │
       ▼
  notifyListeners()
       │
       ▼
  listener1()  listener2()  ...
```

> **注意**：watch 协程在 `NewConsulSubscriber` 返回前即已启动，因此 `AddListener` 可在构造后随时注册，不会丢失首次之后的事件。

### Value() 工作流程

`Value()` 每次调用都会实时读取 Consul KV：

1. 从 Consul KV 读取原始字节数据
2. 如果 key 不存在，返回空字符串（不报错）
3. 使用 viper 按 `Type` 指定的格式解析原始数据
4. 将解析后的 settings 序列化为 JSON 字符串返回

```
Consul KV 原始值 (YAML/JSON/HCL/XML)
        │
        ▼
   viper.ReadConfig()
        │
        ▼
   viper.AllSettings()
        │
        ▼
   json.Marshal() → JSON 字符串
```

> 这意味着无论 Consul KV 中存储的是哪种格式，`Value()` 始终返回 JSON 字符串，与 go-zero `configurator` 的期望一致。

### 与 go-zero configcenter 集成

本模块实现了 go-zero `configcenter.Subscriber` 接口：

```go
// go-zero configcenter.Subscriber 接口定义
type Subscriber interface {
    Value() (string, error)
    AddListener(listener func()) error
}
```

通过 `configurator.MustNewConfigCenter` 将 `ConsulSubscriber` 包装为类型安全的配置中心：

```go
cc := configurator.MustNewConfigCenter[YourConfigType](configurator.Config{
    Type: "yaml", // 配置值格式
}, subscriber)

v, err := cc.GetConfig() // 返回 *YourConfigType
```

### 资源释放

`ConsulSubscriber` 内部运行一个 watch 协程，使用完毕后应调用 `Stop()` 释放资源：

```go
sub, _ := consul.NewConsulSubscriber(conf)
defer sub.Stop()
```

在 go-zero 集成场景下，`configurator` 会在 `AddListener` 中注册回调，但不会主动调用 `Stop()`。如果应用需要优雅关闭，应在 `ServiceGroup` 的停机流程中显式调用。

## 完整示例

### 在 go-zero 中使用

**配置文件**

```yaml
# etc/demoa.yaml
ConfigCenterConsul:
  Host: 127.0.0.1:8500
  Scheme: http
  Key: DemoA.api
  Type: yaml
```

**定义配置结构体**

```go
// internal/config/config.go
package config

import (
    configCenterConsul "github.com/lerity-yao/czt-contrib/configcenter/consul"
    "github.com/zeromicro/go-zero/rest"
)

// BaseConfig 基础配置
// 仅放置配置中心连接信息等高优先级配置
// 其余配置应存放在 Consul KV 中，通过配置中心获取
type BaseConfig struct {
    ConfigCenterConsul configCenterConsul.ConsulConf
}

// Config 项目配置（从 Consul KV 中获取）
type Config struct {
    rest.RestConf
}
```

**订阅配置中心**

```go
// internal/config/subscriber.go
package config

import (
    configCenterConsul "github.com/lerity-yao/czt-contrib/configcenter/consul"
    "github.com/zeromicro/go-zero/core/configcenter/configurator"
)

func SubscriberConsulConfig(b BaseConfig) Config {
    ss := configCenterConsul.MustNewConsulSubscriber(b.ConfigCenterConsul)

    cc := configurator.MustNewConfigCenter[Config](configurator.Config{
        Type: b.ConfigCenterConsul.Type,
    }, ss)

    // 获取初始配置
    v, err := cc.GetConfig()
    if err != nil {
        panic(err)
    }

    // 监听配置变更
    cc.AddListener(func() {
        v, err := cc.GetConfig()
        if err != nil {
            panic(err)
        }
        // 在此处理配置变更后的业务逻辑
        // 注意：K8s 环境下配置变更通常需要重启 Pod，不支持热重载
        println("config changed:", v.Name)
    })

    return v
}
```

**启动服务**

```go
// main.go
package main

import (
    "flag"
    "fmt"

    "demoa/internal/config"
    "demoa/internal/svc"

    "github.com/zeromicro/go-zero/core/conf"
    "github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/demoa.yaml", "the config file")

func main() {
    flag.Parse()

    // 加载基础配置（仅含 Consul 连接信息）
    var b config.BaseConfig
    conf.MustLoad(*configFile, &b)

    // 从 Consul 配置中心订阅完整配置
    c := config.SubscriberConsulConfig(b)

    ctx := svc.NewServiceContext(c)
    serviceGroup := service.NewServiceGroup()
    defer serviceGroup.Stop()

    fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
    serviceGroup.Start()
}
```

### 独立使用

```go
package main

import (
    "fmt"
    "log"

    consul "github.com/lerity-yao/czt-contrib/configcenter/consul"
)

func main() {
    sub, err := consul.NewConsulSubscriber(consul.ConsulConf{
        Host:   "127.0.0.1:8500",
        Scheme: "http",
        Key:    "myapp/config",
        Type:   "yaml",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer sub.Stop()

    // 读取当前配置值（JSON 字符串）
    val, err := sub.Value()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("current config:", val)

    // 注册变更回调
    err = sub.AddListener(func() {
        val, _ := sub.Value()
        fmt.Println("config changed:", val)
    })
    if err != nil {
        log.Fatal(err)
    }

    // 阻塞主协程，等待配置变更
    select {}
}
```

## 更新日志

查看 [CHANGELOG.md](./CHANGELOG.md)