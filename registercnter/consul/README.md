# Consul注册中心使用文档

## 1. 项目介绍

本模块提供了基于Consul的服务注册与发现功能，支持gRPC服务的自动注册、健康检查和动态发现，适用于微服务架构中服务治理场景。

主要功能包括：
- 服务自动注册与注销
- 多种健康检查机制（TTL、HTTP、GRPC）
- 服务健康状态监控与自动恢复
- 基于gRPC的服务发现解析器
- 并发安全的监控管理

## 2. 安装

```bash
go get -u github.com/your-project/bk/czt-contrib/registercnter/consul
```

## 3. 服务注册

### 3.1 基本用法

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
		TTL:       15,                    // TTL健康检查间隔（秒）
		Tag:       []string{"v1", "grpc"}, // 服务标签
	}

	// 创建服务实例
	service := consul.MustNewService(":8080", conf)

	// 注册服务
	if err := service.Register(); err != nil {
		panic(err)
	}
	
	// 注意：在非go-zero环境下，需要手动注销服务
	// defer service.DeregisterService() // 优雅注销
	
	// 在go-zero环境下，不需要手动注销服务，go-zero会通过proc包自动处理服务注销
	
	// 启动您的服务...
}
```

### 3.2 配置选项

`consul.Conf` 结构体包含以下配置项：

| 字段名 | 类型 | 描述 | 默认值 |
|-------|------|------|-------|
| Host | string | Consul服务器地址 | 无（必需） |
| Key | string | 服务名称 | 无（必需） |
| Scheme | string | 连接协议(http/https) | "http" |
| Token | string | Consul访问令牌 | "" |
| CheckType | string | 健康检查类型 | "ttl" |
| TTL | int | TTL健康检查间隔(秒) | 15 |
| CheckTimeout | int | 健康检查超时时间(秒) | 5 |
| ExpiredTTL | int | 服务过期倍数 | 2 |
| Tag | []string | 服务标签 | [] |
| Meta | map[string]string | 服务元数据 | nil |
| CheckHttp | CheckHttpConf | HTTP健康检查配置 | - |

### 3.3 健康检查类型

支持三种健康检查类型：

1. **TTL检查** (`CheckTypeTTL`)
    - 定期更新TTL以保持服务健康状态
    - 适用于需要应用自定义健康逻辑的场景

2. **HTTP检查** (`CheckTypeHttp`)
    - 详细配置选项：
    ```go
    conf := consul.Conf{
        CheckType: consul.CheckTypeHttp,
        CheckHttp: consul.CheckHttpConf{
            Host:        "http://127.0.0.1:8080", // 健康检查基础URL
            Path:        "/health",               // 健康检查路径
            Method:      "GET",                   // HTTP方法
            Header:      map[string][]string{      // 自定义HTTP头
                "Content-Type": {"application/json"},
                "Authorization": {"Bearer token123"},
            },
            Timeout:     5,                        // 请求超时时间（秒）
            Interval:    10,                       // 检查间隔（秒）
            SuccessCode: 200,                      // 认为成功的HTTP状态码
        },
    }
    ```
    - 配置说明：
        - Host: 服务健康检查URL的基础部分，包含协议和地址
        - Path: 健康检查的API路径
        - Method: HTTP请求方法，通常为GET
        - Header: 可选的自定义HTTP请求头
        - Timeout: HTTP请求超时时间
        - Interval: 健康检查执行间隔
        - SuccessCode: 定义什么状态码视为健康状态

3. **GRPC检查** (`CheckTypeGrpc`)
    - 适用于直接检查gRPC服务健康状态

## 4. 服务发现

### 4.1 gRPC客户端使用

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

### 4.2 URL查询参数

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

## 5. 高级用法

### 5.1 自定义监控函数

您可以自定义监控函数来实现特殊的健康检查逻辑。以下是单个和多个监控函数的示例：

#### 单个监控函数示例
```go
import (
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
			// 更新TTL保持健康状态
			if err := cc.UpdateStatus("passing"); err != nil {
				return err
			}
			logx.Infof("Service %s is healthy", cc.GetServiceID())
			
			// 重置重试状态
			state.RetryCount = 0
			state.BackoffTime = 1 * time.Second
			return nil
		}
		
		// 服务不健康，返回错误触发重试逻辑
		return fmt.Errorf("service is unhealthy")
	}
}
```

#### 多个监控函数示例
```go
import (
	"time"
	"github.com/your-project/bk/czt-contrib/registercnter/consul"
	"github.com/zeromicro/go-zero/core/logx"
)

// 系统资源监控函数
func resourceMonitorFunc() consul.MonitorFunc {
	return func(cc *consul.CommonClient, state *consul.MonitorState) error {
		// 检查系统资源（CPU、内存等）
		cpuUsage := getCPUUsage()
		if cpuUsage > 90.0 {
			return fmt.Errorf("high CPU usage: %.2f%%", cpuUsage)
		}
		
		memoryUsage := getMemoryUsage()
		if memoryUsage > 95.0 {
			return fmt.Errorf("high memory usage: %.2f%%", memoryUsage)
		}
		
		return nil // 资源正常
	}
}

// 数据库连接监控函数
func dbMonitorFunc() consul.MonitorFunc {
	return func(cc *consul.CommonClient, state *consul.MonitorState) error {
		// 检查数据库连接
		if !isDatabaseConnected() {
			return fmt.Errorf("database connection failed")
		}
		
		// 检查数据库性能
		dbLatency := getDatabaseLatency()
		if dbLatency > 500*time.Millisecond {
			logx.Warnf("High database latency: %v", dbLatency)
			// 警告但不返回错误
		}
		
		return nil // 数据库正常
	}
}

// 业务状态监控函数
func businessMonitorFunc() consul.MonitorFunc {
	return func(cc *consul.CommonClient, state *consul.MonitorState) error {
		// 检查业务关键指标
		if !checkBusinessHealth() {
			return fmt.Errorf("business health check failed")
		}
		
		return nil // 业务状态正常
	}
}

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
	service.Register()
	
	// 在go-zero环境下，不需要手动注销服务
}
```

所有监控函数会按顺序执行，只要有一个函数返回错误，服务就会被认为不健康。

### 5.2 手动控制服务生命周期

```go
func main() {
	service, _ := consul.NewService(":8080", conf)
	
	// 注册服务
	service.Register()
	
	// 业务逻辑...
	
	// 在非go-zero环境下，需要手动注销服务
	if err := service.Deregister(); err != nil {
		logx.Errorf("Deregister error: %v", err)
	}
	
	// 在go-zero环境下，不需要调用此方法，go-zero会自动处理
}
```

## 6. 最佳实践

### 6.1 服务注册最佳实践

1. **使用Ticker而不是time.Sleep**
    - 库内部已实现基于Ticker的监控机制

2. **设置合理的TTL**
    - TTL值建议设置为15-30秒
    - 服务心跳更新频率应小于TTL值

3. **优雅关闭**
    - 在标准Go环境中，使用`defer service.Deregister()`确保服务注销
    - 在go-zero环境中，不需要手动处理服务注销，go-zero的proc包会自动在程序退出时清理资源
    - 注册系统信号处理确保应用退出时清理资源

### 6.2 服务发现最佳实践

1. **启用负载均衡**
    - 通过`WithDefaultServiceConfig`配置轮询策略

2. **只查询健康服务**
    - 在URL中添加`?healthy=true`参数

3. **使用标签过滤**
    - 利用标签区分不同版本或环境的服务

## 7. API参考

### 7.1 Client接口

```go
type Client interface {
	Register() error                  // 注册服务
	DeregisterService() error         // 注销服务
	IsRegistered() (bool, error)      // 检查服务是否已注册
	GetServiceID() string             // 获取服务ID
	GetRegistration() *api.AgentServiceRegistration // 获取服务注册信息
	UpdateStatus(status string) error // 更新服务状态
	Deregister() error                // 安全注销（停止监控+注销服务）
}
```

### 7.2 服务创建函数

```go
// 创建服务实例
func NewService(listenOn string, c Conf, opts ...ServiceOption) (Client, error)

// 创建服务实例，如果失败则panic
func MustNewService(listenOn string, c Conf, opts ...ServiceOption) Client
```

## 8. 故障排查

### 8.1 常见问题

1. **服务注册失败**
    - 检查Consul服务器地址是否正确
    - 验证Token权限是否足够
    - 检查服务端口是否被占用

2. **服务无法被发现**
    - 确认服务已成功注册到Consul
    - 检查健康检查是否正常
    - 验证URL格式和查询参数是否正确

3. **服务频繁重注册**
    - 检查TTL值是否设置过小
    - 验证网络连接是否稳定
    - 查看监控日志中的错误信息

## 9. 许可证

[MIT License](LICENSE)