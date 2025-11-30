# Consul注册中心使用文档

## 1. 项目介绍

本模块提供了基于Consul的服务注册与发现功能，支持自动注册、健康检查、监控和服务发现，适用于微服务架构中服务治理场景。

主要功能包括：
- 服务自动注册与注销
- 多种健康检查机制（TTL、HTTP、GRPC）
- 服务健康状态监控与自动恢复
- 基于gRPC的服务发现解析器
- 优雅关闭与资源清理
- 容器环境（如Kubernetes）适配

## 2. 项目结构

```
registercnter/consul/
├── README.md        # 使用文档
├── builder.go       # 服务构建器
├── config.go        # 配置结构定义
├── consul_test.go   # 单元测试
├── go.mod           # Go模块文件
├── go.sum           # 依赖校验文件
├── register.go      # 核心注册实现
├── resovler.go      # gRPC服务发现解析器
└── target.go        # 目标服务定义
```

## 3. 安装

```bash
go get -u github.com/your-project/bk/czt-contrib/registercnter/consul
```

## 4. 服务注册

### 4.1 基本用法

```go
import (
	"github.com/your-project/bk/czt-contrib/registercnter/consul"
)

func main() {
	// 配置Consul客户端
	conf := consul.Conf{
		Host:      "127.0.0.1:8500",     // Consul服务器地址
		Key:       "user-service",        // 服务名称
		CheckType: consul.CheckTypeTTL,   // 健康检查类型
		TTL:       20,                    // TTL健康检查间隔（秒）
		Tag:       []string{"v1", "grpc"}, // 服务标签
	}

	// 创建服务实例
	service := consul.MustNewService(":8080", conf)

	// 注册服务
	if err := service.RegisterService(); err != nil {
		panic(err)
	}
	
	// 注意：在非go-zero环境下，需要手动注销服务
	// defer service.DeregisterService() // 优雅注销
	
	// 在go-zero环境下，不需要手动注销服务，go-zero会通过proc包自动处理服务注销
	
	// 启动您的服务...
}
```

### 4.2 配置选项

`consul.Conf` 结构体包含以下配置项：

| 字段名 | 类型 | 描述                                                    | 默认值 |
|-------|------|-------------------------------------------------------|-------|
| Host | string | Consul服务器地址                                           | 无（必需） |
| Key | string | 服务名称                                                  | 无（必需） |
| Scheme | string | 连接协议(http/https)                                      | "http" |
| Token | string | Consul访问令牌                                            | "" |
| CheckType | string | 健康检查类型                                                | "ttl" |
| TTL | int | 健康检查间隔(秒)，TTL/HTTP/GRPC                               | 20 |
| CheckTimeout | int | 当consul需要访问服务的健康检查接口时，即checktype不是 ttl的时候，健康检查超时时间(秒) | 3 |
| ExpiredTTL | int | 服务过期时间系数，基础时间为TTL，过期时间为TTL*ExpiredTTL                           | 3 |
| Tag | []string | 服务标签                                                  | [] |
| Meta | map[string]string | 服务元数据                                                 | nil |
| CheckHttp | CheckHttpConf | HTTP健康检查配置                                            | - |

### 4.3 HTTP健康检查配置

```go
type CheckHttpConf struct {
	Method string // HTTP方法（GET或POST）
	Path   string // 健康检查路径
	Host   string // 健康检查主机
	Port   int    // 健康检查端口
	Scheme string // HTTP协议（http或https）
}
```

### 4.4 健康检查类型

支持三种健康检查类型：

1. **TTL检查** (`CheckTypeTTL`)
    - 定期更新TTL以保持服务健康状态
    - 适用于需要应用自定义健康逻辑的场景

2. **HTTP检查** (`CheckTypeHttp`)
    - 详细配置示例：
    ```go
    conf := consul.Conf{
        CheckType: consul.CheckTypeHttp,
        CheckHttp: consul.CheckHttpConf{
            Method: "GET",
            Path:   "/healthz",
            Host:   "0.0.0.0",
            Port:   6060,
            Scheme: "http",
        },
    }
    ```

3. **GRPC检查** (`CheckTypeGrpc`)
    - 适用于直接检查gRPC服务健康状态
    - 注：当前实现框架已具备，可根据需要进一步完善

## 5. 核心API

### 5.1 Client接口

```go
type Client interface {
	RegisterService() error                  // 注册服务并启动监控
	DeregisterService() error                // 注销服务
	GetServiceID() string                    // 获取服务ID
	GetRegistration() *api.AgentServiceRegistration // 获取服务注册信息
	GetServiceClient() *api.Client           // 获取Consul客户端
}
```

### 5.2 服务创建函数

```go
// 创建服务实例
func NewService(listenOn string, c Conf, opts ...ServiceOption) (Client, error)

// 创建服务实例，如果失败则panic
func MustNewService(listenOn string, c Conf, opts ...ServiceOption) Client
```

## 6. 服务发现

### 6.1 gRPC客户端使用

```go
import (
	"google.golang.org/grpc"
	_ "github.com/your-project/bk/czt-contrib/registercnter/consul" // 自动注册解析器
)

func main() {
	// 使用consul URL创建gRPC连接
	conn, err := grpc.Dial(
		"consul://127.0.0.1:8500/user-service?healthy=true&tag=v1",
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// 创建gRPC客户端并使用
	// ...
}
```

### 6.2 URL查询参数

Consul服务发现URL支持以下查询参数：

| 参数名 | 类型 | 描述 | 默认值 |
|-------|------|------|-------|
| healthy | bool | 是否只查询健康服务 | false |
| tag | string | 服务标签过滤 | "" |
| wait | duration | Consul阻塞查询等待时间 | - |
| timeout | duration | 查询超时时间 | - |
| limit | int | 限制返回服务数量 | 0(无限制) |
| dc | string | 数据中心 | - |
| token | string | Consul访问令牌 | - |

## 7. 高级用法

### 7.1 自定义监控函数

您可以自定义监控函数来实现特殊的健康检查逻辑：

```go
import (
	"fmt"
	"time"
	"github.com/your-project/bk/czt-contrib/registercnter/consul"
	"github.com/zeromicro/go-zero/core/logx"
)

// 自定义监控函数
func customMonitorFunc() consul.MonitorFunc {
	return func(cc *consul.CommonClient, state *consul.MonitorState) error {
		// 自定义健康检查逻辑
		isHealthy := checkMyServiceHealth()
		
		if isHealthy {
			// 重置重试状态
			state.RetryCount = 0
			state.BackoffTime = 1 * time.Second
			state.Ticker.Reset(state.OriginalTTL)
			logx.Infof("Service %s is healthy", cc.GetServiceID())
			return nil
		}
		
		// 服务不健康，返回错误触发重试逻辑
		return fmt.Errorf("service is unhealthy")
	}
}

func main() {
	// 使用自定义监控函数
	service, _ := consul.NewService(":8080", conf, 
		consul.WithMonitorFuncs(customMonitorFunc()),
	)
	
	// 注册服务
	service.RegisterService()
}
```

### 7.2 多监控函数支持

您可以注册多个监控函数，每个函数负责不同的健康检查维度：

```go
import (
	"github.com/your-project/bk/czt-contrib/registercnter/consul"
)

func main() {
	// 使用多个自定义监控函数
	service, _ := consul.NewService(":8080", conf, 
		consul.WithMonitorFuncs(
			resourceMonitorFunc(),  // 监控系统资源
			dbMonitorFunc(),        // 监控数据库连接
			businessMonitorFunc(),  // 监控业务状态
		),
	)
	
	// 注册服务
	service.RegisterService()
}
```

## 8. 自动恢复机制

当服务健康检查失败时，系统会自动尝试重新注册：

- 最大重试次数：5次
- 初始退避时间：1秒
- 最大退避时间：30秒
- 采用指数退避策略

## 9. 容器环境适配

模块会自动检测容器环境，优先使用以下方式获取服务地址：

1. 检查`POD_IP`环境变量（Kubernetes容器环境）
2. 使用系统内部IP
3. 回退到配置的监听地址

## 10. 优雅关闭

通过`proc.AddShutdownListener`机制，在程序退出时自动：

1. 停止所有监控协程
2. 注销服务
3. 清理资源

## 11. 最佳实践

### 11.1 服务注册最佳实践

1. **设置合理的TTL**
    - TTL值建议设置为15-30秒
    - 系统会自动以TTL-1秒的频率发送心跳

2. **优雅关闭**
    - 在标准Go环境中，使用`defer service.DeregisterService()`确保服务注销
    - 在go-zero环境中，不需要手动处理服务注销

3. **合理设置健康检查**
    - HTTP检查适用于有Web接口的服务
    - TTL检查适用于需要自定义健康逻辑的场景

### 11.2 服务发现最佳实践

1. **启用负载均衡**
    - 通过`WithDefaultServiceConfig`配置轮询策略

2. **只查询健康服务**
    - 在URL中添加`?healthy=true`参数

3. **使用标签过滤**
    - 利用标签区分不同版本或环境的服务

## 12. 故障排查

### 12.1 常见问题

1. **服务注册失败**
    - 检查Consul服务器地址是否正确
    - 验证Token权限是否足够
    - 检查服务端口是否被占用

2. **健康检查失败**
    - TTL模式：检查网络连接是否稳定
    - HTTP模式：验证健康检查端点是否正确配置并返回200状态

3. **服务自动注销**
    - 检查TTL设置是否合理
    - 查看日志中的错误信息
    - 验证系统时间是否同步

4. **服务发现问题**
    - 检查Consul URL格式是否正确
    - 验证服务是否已经正确注册到Consul
    - 确认查询参数设置是否合适

### 12.2 日志排查

模块使用`github.com/zeromicro/go-zero/core/logx`记录日志，可通过配置logx查看详细日志信息。

## 13. 依赖项

- github.com/hashicorp/consul/api
- github.com/zeromicro/go-zero/core/logx
- github.com/zeromicro/go-zero/core/netx
- github.com/zeromicro/go-zero/core/proc
- google.golang.org/grpc

## 14. 许可证

[MIT License](LICENSE)
        