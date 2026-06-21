# consul

English | [中文](./readme-cn.md)

[![codecov](https://codecov.io/gh/lerity-yao/czt-contrib/branch/main/graph/badge.svg?flag=registercenter-consul)](https://codecov.io/gh/lerity-yao/czt-contrib)

Service registration and discovery module based on [Consul](https://developer.hashicorp.com/consul) and [go-zero](https://github.com/zeromicro/go-zero), supporting automatic registration, health checks (TTL / HTTP / gRPC), automatic recovery, and a gRPC service discovery resolver.

## Features

- 📋 **Auto Registration & Deregistration** — Service is automatically registered on startup, and deregistered on process exit via `proc.AddShutdownListener`
- 💓 **Multiple Health Checks** — Supports TTL, HTTP, and gRPC health check mechanisms
- 🔄 **Automatic Recovery** — Automatically retries registration on health check failure with exponential backoff
- 🔍 **gRPC Service Discovery** — Built-in `consul://` scheme resolver, auto-registered via `init()`, supporting blocking queries and tag filtering
- 🐳 **Container Environment Adaptation** — Automatically detects `POD_IP` environment variable (Kubernetes), falls back to internal IP
- 🔧 **Extensible Monitoring** — Inject custom monitor functions via `WithMonitorFuncs`

## Installation

```bash
go get github.com/lerity-yao/czt-contrib/registercenter/consul@latest
```

## Configuration Parameters

### Conf

| Parameter | Type | Required | Default | Description |
|--------|------|:----:|--------|------|
| `Host` | string | Yes | - | Consul server address, format `host:port`, e.g. `127.0.0.1:8500` |
| `Key` | string | Yes | - | Service name, e.g. `user-service` |
| `Scheme` | string | No | `http` | Connection protocol, options: `http` / `https` |
| `Token` | string | No | `""` | Consul ACL access token |
| `Tag` | []string | No | `[]` | Service tag list |
| `Meta` | map[string]string | No | `nil` | Service metadata |
| `TTL` | int | No | `20` | Health check interval (seconds). In TTL mode, this is the heartbeat interval; in HTTP / gRPC mode, this is the interval for Consul server-initiated checks |
| `ExpiredTTL` | int | No | `3` | Service deregistration multiplier. Actual deregistration time is `TTL * ExpiredTTL` seconds |
| `CheckTimeout` | int | No | `3` | Health check timeout (seconds). Only effective for HTTP / gRPC mode (not used in TTL mode) |
| `CheckType` | string | No | `ttl` | Health check type, options: `ttl` / `http` / `grpc` |
| `CheckHttp` | [CheckHttpConf](#checkhttpconf) | No | - | HTTP health check configuration, effective when `CheckType` is `http` |
| `CheckGrpc` | [CheckGrpcConf](#checkgrpcconf) | No | - | gRPC health check configuration, effective when `CheckType` is `grpc` |

> `Conf.Validate()` is automatically called when invoking `NewService` to validate the above fields.

### CheckHttpConf

| Parameter | Type | Default | Description |
|--------|------|--------|------|
| `Method` | string | `GET` | HTTP method, options: `GET` / `POST` |
| `Path` | string | `/healthz` | Health check path |
| `Host` | string | `0.0.0.0` | Health check host address |
| `Port` | int | `6060` | Health check port |
| `Scheme` | string | `http` | HTTP protocol, options: `http` / `https` |

### CheckGrpcConf

| Parameter | Type | Default | Description |
|--------|------|--------|------|
| `TLSServerName` | string | `""` | TLS server name (optional), used for TLS connection verification |
| `TLSSkipVerify` | bool | `true` | Whether to skip TLS verification |
| `GRPCUseTLS` | bool | `false` | Whether to use TLS connection |

## API Reference

### Constructors

| Function | Signature | Description |
|------|------|------|
| `MustNewService` | `func MustNewService(listenOn string, c Conf, opts ...ServiceOption) Client` | Create service instance, panics on validation failure |
| `NewService` | `func NewService(listenOn string, c Conf, opts ...ServiceOption) (Client, error)` | Create service instance, returns error on validation failure |

> `listenOn` is the service listen address, e.g. `:8080` or `0.0.0.0:8080`. The module automatically resolves it to the actual reachable IP:Port.

### ServiceOption

| Option | Parameter | Description |
|--------|------|------|
| `WithMonitorFuncs` | `funcs ...MonitorFunc` | Inject custom monitor functions. If not provided, the default monitor function is automatically selected based on `CheckType` |

### Client Interface Methods

| Method | Signature | Description |
|------|------|------|
| `RegisterService` | `RegisterService() error` | Register service and start health monitoring, auto-register graceful shutdown callback |
| `DeregisterService` | `DeregisterService() error` | Deregister service and stop all monitor goroutines |
| `GetServiceID` | `GetServiceID() string` | Get service ID, format is `Key-Host-Port` |
| `GetRegistration` | `GetRegistration() *api.AgentServiceRegistration` | Get service registration info |
| `GetServiceClient` | `GetServiceClient() *api.Client` | Get Consul API client instance |

### Monitor Functions

| Function | Signature | Description |
|------|------|------|
| `TTLCheckMonitorFunc` | `func TTLCheckMonitorFunc() MonitorFunc` | Default monitor function for TTL health checks, periodically calls `UpdateTTL` to update heartbeat |
| `HttpCheckMonitorFunc` | `func HttpCheckMonitorFunc() MonitorFunc` | Default monitor function for HTTP / gRPC health checks, periodically queries service health status |
| `TTLMonitorLogic` | `func TTLMonitorLogic(cc *CommonClient, state *MonitorState) error` | TTL monitor logic, includes automatic registration retry |
| `HttpMonitorLogic` | `func HttpMonitorLogic(cc *CommonClient, state *MonitorState) error` | HTTP / gRPC monitor logic, includes automatic registration retry |

> **Default monitor function selection rule**: When `CheckType` is `ttl`, `TTLCheckMonitorFunc()` is used; when `http` or `grpc`, `HttpCheckMonitorFunc()` is used.

### Public Types

| Type | Definition | Description |
|------|------|------|
| `MonitorFunc` | `func(cc *CommonClient, stopChan <-chan struct{})` | Monitor function signature, receives `CommonClient` and stop channel |
| `ServiceOption` | `func(*CommonClient)` | Service option function signature |
| `MonitorState` | `struct{...}` | Monitor state, includes retry count, backoff time, Ticker, etc., provides `Close()` method |

### Constants

| Constant | Value | Description |
|------|------|------|
| `CheckTypeTTL` | `"ttl"` | TTL health check type |
| `CheckTypeHttp` | `"http"` | HTTP health check type |
| `CheckTypeGrpc` | `"grpc"` | gRPC health check type |

## Advanced Guide

### Health Check Mechanism

Three health check types are supported, determined by `Conf.CheckType`:

#### TTL Check (`ttl`)

- **Mechanism**: Service periodically sends `UpdateTTL` heartbeat to Consul to maintain healthy status
- **Heartbeat frequency**: `TTL - 1` seconds (minimum 1 second)
- **Use case**: Scenarios requiring custom application health logic
- Consul will automatically deregister the service if no heartbeat is received within `TTL * ExpiredTTL` seconds

#### HTTP Check (`http`)

- **Mechanism**: Consul server periodically sends HTTP requests to the service's health check endpoint at `TTL` second intervals
- **Timeout**: Controlled by `CheckTimeout`
- **Use case**: Services with web interfaces
- After enabling health checks in go-zero, `host:6060/healthz` endpoint is available by default

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

#### gRPC Check (`grpc`)

- **Mechanism**: Consul server periodically sends gRPC health check requests to the service at `TTL` second intervals
- **Timeout**: Controlled by `CheckTimeout`
- **Use case**: gRPC services
- After enabling rpc service in go-zero, `grpc.health.v1.Health/Check` endpoint is available by default

```go
conf := consul.Conf{
    Host:      "127.0.0.1:8500",
    Key:       "user-service",
    CheckType: consul.CheckTypeGrpc,
    TTL:       20,
    CheckTimeout: 5,
    CheckGrpc: consul.CheckGrpcConf{
        TLSServerName: "example.com", // optional
        TLSSkipVerify: true,
        GRPCUseTLS:    false,
    },
}
```

> When using gRPC check, the service must implement the standard health check interface (`grpc.health.v1.Health`).

### Automatic Recovery Mechanism

When a health check fails (TTL update failure or HTTP / gRPC health status is not `passing`), the monitor goroutine automatically attempts to re-register the service:

| Parameter | Value | Description |
|------|------|------|
| Max retry attempts | 5 | Counter resets and continues retrying after exceeding |
| Initial backoff time | 1 second | Wait time for first retry |
| Max backoff time | 30 seconds | Backoff time upper limit |
| Backoff strategy | Exponential backoff | `backoff * 2`, capped at upper limit |

Retry flow:
1. Health check failure
2. Check if current service health status is `passing`
3. If not and max retries not reached, call `registerServiceWithPassingHealth()` to re-register
4. On retry failure, increase backoff time; on retry success, reset counter and restore original heartbeat frequency

### Container Environment Adaptation

The module automatically resolves the service's reachable address via `figureOutListenOn`:

1. Check `POD_IP` environment variable (injected by Kubernetes container environment)
2. Use go-zero `netx.InternalIp()` to get system internal IP
3. Fall back to configured listen address

> Address resolution is triggered when `listenOn` host is `0.0.0.0`. Non-`0.0.0.0` addresses remain unchanged.

### Service Discovery URL Parameters

The `consul://` scheme gRPC resolver is auto-registered via `init()`. URL format:

```
consul://[user:passwd]@host/service?param=value
```

| Parameter | Type | Default | Description |
|------|------|--------|------|
| `healthy` | bool | `false` | Whether to query only healthy services |
| `tag` | string | `""` | Service tag filter |
| `wait` | duration | - | Consul blocking query wait time |
| `timeout` | duration | - | Query timeout |
| `max-backoff` | duration | `1s` | Max backoff time on fetch failure |
| `near` | string | `_agent` | Sort by distance for nearest access |
| `limit` | int | `0` | Limit number of returned services (0 = no limit) |
| `insecure` | bool | `false` | Whether to skip TLS verification |
| `token` | string | `""` | Consul ACL access token |
| `dc` | string | `""` | Datacenter |
| `allow-stale` | bool | `false` | Whether to allow stale data |
| `require-consistent` | bool | `false` | Whether to require consistent read |

### Graceful Shutdown

`RegisterService()` internally registers a shutdown callback via `proc.AddShutdownListener`, which is automatically executed on program exit:

1. Stop all monitor goroutines (close stop channel)
2. Call `ServiceDeregister` to deregister the service
3. Log deregistration result

> In go-zero environments, manual deregistration is not needed; in non-go-zero environments, use `defer service.DeregisterService()` to ensure deregistration.

## Complete Examples

### Using in go-zero

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

### Standalone Usage

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
    // In non-go-zero environments, manual deregistration is required
    defer service.DeregisterService()

    fmt.Println("service registered:", service.GetServiceID())
    // Start your service...
}
```

### Custom Monitor Functions

```go
conf := consul.Conf{
    Host:      "127.0.0.1:8500",
    Key:       "user-service",
    CheckType: consul.CheckTypeTTL,
    TTL:       20,
}

// Custom monitor function
func customMonitorFunc() consul.MonitorFunc {
    return func(cc *consul.CommonClient, stopCh <-chan struct{}) {
        // Your custom monitor logic
    }
}

service, _ := consul.NewService(":8080", conf,
    consul.WithMonitorFuncs(
        consul.TTLCheckMonitorFunc(),   // Keep default TTL monitoring
        customMonitorFunc(),             // Append custom monitoring
    ),
)

service.RegisterService()
```

### Service Discovery (gRPC Client)

```go
import (
    "google.golang.org/grpc"
    _ "github.com/lerity-yao/czt-contrib/registercenter/consul" // Auto-register resolver
)

func main() {
    // Create gRPC connection using consul URL
    conn, err := grpc.Dial(
        "consul://127.0.0.1:8500/user-service?healthy=true&tag=v1",
        grpc.WithInsecure(),
        grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
    )
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    // Create gRPC client and use it
    // ...
}
```

## Changelog

See [CHANGELOG.md](./CHANGELOG.md)
# Consul Service Registry Documentation

[中文](./readme-cn.md)

[![codecov](https://codecov.io/gh/lerity-yao/czt-contrib/branch/main/graph/badge.svg?flag=registercenter-consul)](https://codecov.io/gh/lerity-yao/czt-contrib)

## 1. Overview

This module provides Consul-based service registration and discovery, supporting automatic registration, health checks, monitoring, and service discovery. It is suitable for service governance in microservice architectures.

Key features:
- Automatic service registration and deregistration
- Multiple health check mechanisms (TTL, HTTP, GRPC)
- Service health monitoring and automatic recovery
- gRPC-based service discovery resolver
- Graceful shutdown and resource cleanup
- Container environment (e.g., Kubernetes) adaptation

## 2. Project Structure
```
registercnter/consul/ 
├── README.md # Documentation 
├── builder.go # Service builder 
├── config.go # Configuration struct definitions 
├── consul_test.go # Unit tests
├── go.mod # Go module file 
├── go.sum # Dependency checksum file 
├── register.go # Core registration implementation 
├── resovler.go # gRPC service discovery resolver 
└── target.go # Target service definition
```

## 3. Installation

```bash
go get -u github.com/lerity-yao/czt-contrib/registercenter/consul
```

## 4. Service Registration

### 4.1 Basic Usage

```go
import (
	"github.com/your-project/bk/czt-contrib/registercnter/consul"
)

func main() {
	// Configure the Consul client
	conf := consul.Conf{
		Host:      "127.0.0.1:8500",     // Consul server address
		Key:       "user-service",        // Service name
		CheckType: consul.CheckTypeTTL,   // Health check type
		TTL:       20,                    // TTL health check interval (seconds)
		Tag:       []string{"v1", "grpc"}, // Service tags
	}

	// Create service instance
	service := consul.MustNewService(":8080", conf)

	// Register the service
	if err := service.RegisterService(); err != nil {
		panic(err)
	}
	
	// Note: In non-go-zero environments, you need to deregister the service manually
	// defer service.DeregisterService() // graceful deregistration
	
	// In go-zero environments, manual deregistration is not needed;
	// go-zero handles it automatically via the proc package
	
	// Start your service...
}
```

### 4.2 Configuration Options

The `consul.Conf` struct contains the following fields:

| Field        | Type              | Description                                                                                          | Default                        |
|--------------|-------------------|------------------------------------------------------------------------------------------------------|--------------------------------|
| Host         | string            | Consul server address                                                                                | required                       |
| Key          | string            | Service name                                                                                         | required                       |
| Scheme       | string            | Connection protocol (http/https)                                                                     | "http"                         |
| Token        | string            | Consul access token                                                                                  | ""                             |
| CheckType    | string            | Health check type                                                                                    | "ttl"; options: ttl, http, grpc |
| TTL          | int               | Health check interval (seconds) for TTL/HTTP/GRPC                                                   | 20                             |
| CheckTimeout | int               | Health check timeout (seconds) when consul accesses the service health endpoint (non-TTL check types) | 3                              |
| ExpiredTTL   | int               | Service expiry multiplier; expiry time = TTL * ExpiredTTL                                            | 3                              |
| Tag          | []string          | Service tags                                                                                         | []                             |
| Meta         | map[string]string | Service metadata                                                                                     | nil                            |
| CheckHttp    | CheckHttpConf     | HTTP health check config; effective when CheckType is http                                           | -                              |
| CheckGrpc    | CheckGrpcConf     | GRPC health check config; effective when CheckType is grpc                                           | -                              |

### 4.3 HTTP Health Check Configuration

```go
type CheckHttpConf struct {
	Method string // HTTP method (GET or POST)
	Path   string // Health check path
	Host   string // Health check host
	Port   int    // Health check port
	Scheme string // HTTP scheme (http or https)
}
```

### 4.4 GRPC Health Check Configuration

```go
type CheckGrpcConf struct {
	TLSServerName string // TLS server name (optional)
	TLSSkipVerify bool   // Whether to skip TLS verification, default true
	GRPCUseTLS    bool   // Whether to use TLS, default false
}
```

### 4.5 Health Check Types

Three health check types are supported:

1. **TTL Check** (`CheckTypeTTL`)
    - Periodically updates the TTL to maintain service health status
    - Suitable for scenarios requiring custom health logic in the application

2. **HTTP Check** (`CheckTypeHttp`)
    - Suitable for directly checking the health of API services
    - The `http` check is initiated by the `consul` server-side toward the service on a scheduled basis
    - The service must expose a health-check endpoint; after enabling health checks in `go-zero`, a `host:6060/healthz` endpoint is available by default
    - Detailed configuration example:
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

3. **GRPC Check** (`CheckTypeGrpc`)
    - Suitable for directly checking the health of gRPC services
    - The `grpc` check is initiated by the `consul` server-side toward the service on a scheduled basis
    - The service must expose a health-check endpoint; after enabling an rpc service in `go-zero`, a `grpc.health.v1.Health/Check` endpoint is available by default
    - Detailed configuration example:
    ```go
    conf := consul.Conf{
        CheckType: consul.CheckTypeGrpc,
        TTL:       20,  // Health check interval
        CheckTimeout: 5, // Health check timeout
        CheckGrpc: consul.CheckGrpcConf{
            TLSServerName: "example.com", // Optional, used for TLS connection verification
            TLSSkipVerify: true,          // Whether to skip TLS verification, default true
            GRPCUseTLS:    false,         // Whether to use TLS connection, default false
        },
    }
    ```
    - Note: When using the GRPC check, your gRPC service must implement the standard health check service interface (`grpc.health.v1.Health`)

## 5. Core API

### 5.1 Client Interface

```go
type Client interface {
	RegisterService() error                  // Register service and start monitoring
	DeregisterService() error                // Deregister service
	GetServiceID() string                    // Get service ID
	GetRegistration() *api.AgentServiceRegistration // Get service registration info
	GetServiceClient() *api.Client           // Get Consul client
}
```

### 5.2 Service Constructor Functions

```go
// Create a service instance
func NewService(listenOn string, c Conf, opts ...ServiceOption) (Client, error)

// Create a service instance; panics on failure
func MustNewService(listenOn string, c Conf, opts ...ServiceOption) Client
```

## 6. Service Discovery

### 6.1 gRPC Client Usage

```go
import (
	"google.golang.org/grpc"
	_ "github.com/lerity-yao/czt-contrib/registercenter/consul" // Auto-register resolver
)

func main() {
	// Create a gRPC connection using a consul URL
	conn, err := grpc.Dial(
		"consul://127.0.0.1:8500/user-service?healthy=true&tag=v1",
		grpc.WithInsecure(),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
	)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Create and use a gRPC client
	// ...
}
```

### 6.2 URL Query Parameters

The Consul service discovery URL supports the following query parameters:

| Parameter | Type     | Description                          | Default    |
|-----------|----------|--------------------------------------|------------|
| healthy   | bool     | Whether to query only healthy services | false    |
| tag       | string   | Service tag filter                   | ""         |
| wait      | duration | Consul blocking query wait time      | -          |
| timeout   | duration | Query timeout                        | -          |
| limit     | int      | Limit the number of returned services | 0 (no limit) |
| dc        | string   | Datacenter                           | -          |
| token     | string   | Consul access token                  | -          |

## 7. Advanced Usage

### 7.1 Custom Monitor Functions

You can define custom monitor functions to implement special health check logic:

```go
import (
	"fmt"
	"time"
	"github.com/lerity-yao/czt-contrib/registercenter/consul"
	"github.com/zeromicro/go-zero/core/logx"
)

// Custom monitor function
func customMonitorFunc() consul.MonitorFunc {
	return func(cc *CommonClient, stopCh <-chan struct{}) {
		// todo your logic
	}
}

func main() {
	// Use a custom monitor function
	service, _ := consul.NewService(":8080", conf, 
		consul.WithMonitorFuncs(customMonitorFunc()),
	)
	
	// Register the service
	service.RegisterService()
}
```

### 7.2 Multiple Monitor Functions

You can register multiple monitor functions, each responsible for a different health check dimension:

```go
import (
	"github.com/lerity-yao/czt-contrib/registercenter/consul"
)

func main() {
	// Use multiple custom monitor functions
	service, _ := consul.NewService(":8080", conf, 
		consul.WithMonitorFuncs(
			resourceMonitorFunc(),  // Monitor system resources
			dbMonitorFunc(),        // Monitor database connections
			businessMonitorFunc(),  // Monitor business status
		),
	)
	
	// Register the service
	service.RegisterService()
}
```

## 8. Automatic Recovery

When a service health check fails, the system automatically attempts to re-register:

- Maximum retry attempts: 5
- Initial backoff time: 1 second
- Maximum backoff time: 30 seconds
- Uses exponential backoff strategy

## 9. Container Environment Adaptation

The module automatically detects container environments and prioritizes the following methods for obtaining the service address:

1. Check the `POD_IP` environment variable (Kubernetes container environment)
2. Use the system's internal IP
3. Fall back to the configured listen address

## 10. Graceful Shutdown

Via the `proc.AddShutdownListener` mechanism, the following actions are performed automatically on program exit:

1. Stop all monitor goroutines
2. Deregister the service
3. Clean up resources

## 11. Best Practices

### 11.1 Service Registration Best Practices

1. **Set a reasonable TTL**
    - Recommended TTL: 15–30 seconds
    - The system automatically sends heartbeats at a frequency of TTL-1 seconds

2. **Graceful shutdown**
    - In standard Go environments, use `defer service.DeregisterService()` to ensure deregistration
    - In go-zero environments, manual deregistration is not required

3. **Configure health checks appropriately**
    - HTTP checks are suitable for services with web interfaces
    - TTL checks are suitable for scenarios requiring custom health logic
    - GRPC checks are suitable for gRPC services that implement the standard health check interface

### 11.2 Service Discovery Best Practices

1. **Enable load balancing**
    - Configure the round-robin policy via `WithDefaultServiceConfig`

2. **Query only healthy services**
    - Add `?healthy=true` to the URL

3. **Use tag filtering**
    - Use tags to distinguish between different versions or environments of a service

## 12. Troubleshooting

### 12.1 Common Issues

1. **Service registration failure**
    - Check that the Consul server address is correct
    - Verify that the Token has sufficient permissions
    - Check that the service port is not already in use

2. **Health check failure**
    - TTL mode: Check that the network connection is stable
    - HTTP mode: Verify that the health check endpoint is correctly configured and returns a 200 status
    - GRPC mode: Ensure the service implements the standard health check interface

3. **Service auto-deregistration**
    - Check that the TTL setting is reasonable
    - Review error messages in the logs
    - Verify that the system clock is synchronized

4. **Service discovery issues**
    - Check that the Consul URL format is correct
    - Verify that the service has been correctly registered with Consul
    - Confirm that the query parameters are set appropriately

### 12.2 Log Diagnostics

The module uses `github.com/zeromicro/go-zero/core/logx` for logging. Configure logx to view detailed log information.

## 13. Dependencies

- github.com/hashicorp/consul/api
- github.com/zeromicro/go-zero/core/logx
- github.com/zeromicro/go-zero/core/netx
- github.com/zeromicro/go-zero/core/proc
- google.golang.org/grpc

## 14. License

[MIT License](LICENSE)

## 15. Changelog

See [CHANGELOG.md](./CHANGELOG.md)
