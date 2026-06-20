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
