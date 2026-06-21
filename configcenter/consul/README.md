# configcenter/consul

English | [中文](./readme-cn.md)

A go-zero configuration center subscription module based on [Consul KV](https://developer.hashicorp.com/consul/api-docs/kv). It automatically watches KV changes and notifies the business layer, implementing the go-zero `configcenter.Subscriber` interface.

## Features

- 🔍 **Automatic Change Watch** — Based on Consul blocking query long polling, KV changes trigger callbacks in real time
- 📄 **Multiple Format Support** — Supports YAML, JSON, HCL, and XML configuration formats, unified output as JSON
- 🔌 **Seamless go-zero Integration** — Implements the `configcenter.Subscriber` interface, works out of the box with `configurator.MustNewConfigCenter`
- 🔒 **TLS and ACL** — Supports Consul Token authentication and TLS encrypted connections

## Installation

```bash
go get github.com/lerity-yao/czt-contrib/configcenter/consul@v0.1.2
```

## Configuration Parameters

### Conf

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `Host` | string | No | - | Consul address, in `host:port` format, e.g. `127.0.0.1:8500` |
| `Scheme` | string | No | `http` | Consul address protocol, `http` or `https` |
| `PathPrefix` | string | No | - | Consul API path prefix |
| `Datacenter` | string | No | - | Consul datacenter name |
| `Token` | string | No | - | Consul ACL Token |
| `TLSConfig` | `api.TLSConfig` | No | - | Consul TLS connection configuration |
| `Key` | string | No | - | Consul KV path, i.e. the key of the configuration in KV, e.g. `DemoA.api` |
| `Type` | string | No | `yaml` | Configuration value format, optional values: `yaml`, `hcl`, `json`, `xml` |

> `ConsulConf` is a type alias for `Conf` (`type ConsulConf Conf`), the two are equivalent; `ConsulConf` is recommended.

## API Reference

### Constructors

| Function | Signature | Description |
|----------|-----------|-------------|
| `MustNewConsulSubscriber` | `func MustNewConsulSubscriber(conf ConsulConf) *ConsulSubscriber` | Creates a Subscriber, panics on failure |
| `NewConsulSubscriber` | `func NewConsulSubscriber(conf ConsulConf) (*ConsulSubscriber, error)` | Creates a Subscriber, returns an error on failure |

### ConsulSubscriber Methods

| Method | Signature | Description |
|--------|-----------|-------------|
| `Value` | `func (s *ConsulSubscriber) Value() (string, error)` | Reads the current value from Consul KV, parses it, and returns it as a JSON string |
| `AddListener` | `func (s *ConsulSubscriber) AddListener(listener func()) error` | Registers a change callback, automatically triggered when the KV changes |
| `Stop` | `func (s *ConsulSubscriber) Stop()` | Stops the background watch goroutine |

> `Value()` and `AddListener()` together implement the go-zero `configcenter.Subscriber` interface.

### ConsulSubscriber Exported Fields

| Field | Type | Description |
|-------|------|-------------|
| `Path` | string | Consul KV path, assigned from `Conf.Key` during creation |
| `Type` | string | Configuration format, assigned from `Conf.Type` during creation |

## Advanced Guide

### Watch Mechanism

`ConsulSubscriber` automatically starts a background watch goroutine upon creation, implementing change listening based on the Consul KV **blocking query** mechanism:

1. **First Request**: Issues a KV Get request and records `X-Consul-Index` (i.e. `LastIndex`)
2. **Long Polling**: Subsequent requests carry `WaitIndex=LastIndex`; Consul blocks until the value changes (no timeout by default)
3. **Change Detection**: When the KV value is modified, Consul returns the new value and a larger `LastIndex`
4. **Notify Callbacks**: All registered listeners are triggered when `LastIndex` increases
5. **Error Retry**: Waits 1 second before retrying after a request failure

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

> **Note**: The watch goroutine is started before `NewConsulSubscriber` returns, so `AddListener` can be registered at any time after construction without missing events after the first one.

### Value() Workflow

`Value()` reads from Consul KV in real time on every call:

1. Reads the raw bytes from Consul KV
2. Returns an empty string (without error) if the key does not exist
3. Parses the raw data using viper according to the format specified by `Type`
4. Serializes the parsed settings into a JSON string and returns it

```
Consul KV raw value (YAML/JSON/HCL/XML)
        │
        ▼
   viper.ReadConfig()
        │
        ▼
   viper.AllSettings()
        │
        ▼
   json.Marshal() → JSON string
```

> This means no matter which format is stored in Consul KV, `Value()` always returns a JSON string, consistent with go-zero `configurator` expectations.

### Integration with go-zero configcenter

This module implements the go-zero `configcenter.Subscriber` interface:

```go
// go-zero configcenter.Subscriber interface definition
type Subscriber interface {
    Value() (string, error)
    AddListener(listener func()) error
}
```

Use `configurator.MustNewConfigCenter` to wrap `ConsulSubscriber` as a type-safe configuration center:

```go
cc := configurator.MustNewConfigCenter[YourConfigType](configurator.Config{
    Type: "yaml", // configuration value format
}, subscriber)

v, err := cc.GetConfig() // returns *YourConfigType
```

### Resource Release

`ConsulSubscriber` runs an internal watch goroutine; call `Stop()` to release resources after use:

```go
sub, _ := consul.NewConsulSubscriber(conf)
defer sub.Stop()
```

In the go-zero integration scenario, `configurator` registers a callback in `AddListener`, but does not actively call `Stop()`. If the application needs graceful shutdown, `Stop()` should be called explicitly in the `ServiceGroup` shutdown process.

## Complete Examples

### Using with go-zero

**Configuration File**

```yaml
# etc/demoa.yaml
ConfigCenterConsul:
  Host: 127.0.0.1:8500
  Scheme: http
  Key: DemoA.api
  Type: yaml
```

**Define Configuration Struct**

```go
// internal/config/config.go
package config

import (
    configCenterConsul "github.com/lerity-yao/czt-contrib/configcenter/consul"
    "github.com/zeromicro/go-zero/rest"
)

// BaseConfig basic configuration
// Only holds high-priority configuration such as the config center connection info
// Other configurations should be stored in Consul KV and obtained via the config center
type BaseConfig struct {
    ConfigCenterConsul configCenterConsul.ConsulConf
}

// Config project configuration (obtained from Consul KV)
type Config struct {
    rest.RestConf
}
```

**Subscribe to the Config Center**

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

    // Get initial configuration
    v, err := cc.GetConfig()
    if err != nil {
        panic(err)
    }

    // Listen for configuration changes
    cc.AddListener(func() {
        v, err := cc.GetConfig()
        if err != nil {
            panic(err)
        }
        // Handle business logic after configuration changes here
        // Note: In K8s environments, configuration changes usually require restarting the Pod; hot reload is not supported
        println("config changed:", v.Name)
    })

    return v
}
```

**Start the Service**

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

    // Load basic configuration (only contains Consul connection info)
    var b config.BaseConfig
    conf.MustLoad(*configFile, &b)

    // Subscribe to the full configuration from the Consul config center
    c := config.SubscriberConsulConfig(b)

    ctx := svc.NewServiceContext(c)
    serviceGroup := service.NewServiceGroup()
    defer serviceGroup.Stop()

    fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
    serviceGroup.Start()
}
```

### Standalone Use

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

    // Read the current configuration value (JSON string)
    val, err := sub.Value()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println("current config:", val)

    // Register change callback
    err = sub.AddListener(func() {
        val, _ := sub.Value()
        fmt.Println("config changed:", val)
    })
    if err != nil {
        log.Fatal(err)
    }

    // Block the main goroutine, waiting for configuration changes
    select {}
}
```

## Changelog

See [CHANGELOG.md](./CHANGELOG.md)
# configcenter/consul

[中文](./readme-cn.md)

A `consul` configuration center integration for `go-zero`.

## Usage

Add the following configuration to your config file:

```yaml
ConfigCenterConsul:
  Host: 127.0.0.1:8500
  Scheme: http
  Key: DemoA.api
```

```go

package config

import (
	configCenterConsul "github.com/lerity-yao/czt-contrib/configcenter/consul"
	"github.com/zeromicro/go-zero/core/configcenter"
	"github.com/zeromicro/go-zero/rest"
)

// BaseConfig holds the base configuration.
// BaseConfig currently only contains the config center configuration, allowing the project
// to connect to the config center and retrieve other config files with only the config center settings.
// If a config item has higher priority than the config center, it may be placed here;
// otherwise, it must go into Config instead of BaseConfig.
type BaseConfig struct {
	ConfigCenterConsul configCenterConsul.ConsulConf // Config center configuration
}

// Config holds the project configuration.
type Config struct {
	rest.RestConf
}

// SubscriberConsulConfig subscribes to the consul config center.
// It supports monitoring configuration changes but does not support hot-reloading.
// In Kubernetes, configuration changes require a pod restart, so hot-reload is not supported.
func SubscriberConsulConfig(b BaseConfig) Config {

	ss := configCenterConsul.MustNewConsulSubscriber(b.ConfigCenterConsul)
	// Create configurator
	cc := configurator.MustNewConfigCenter[Config](configurator.Config{
		Type: "yaml", // Config value type: json, yaml, toml
	}, ss)

	// Get config
	// Note: if the config changes, calling this will always return the latest config
	v, err := cc.GetConfig()
	if err != nil {
		panic(err)
	}
	cc.AddListener(func() {
		v, err := cc.GetConfig()
		if err != nil {
			panic(err)
		}
		// Write the operations to perform after a config change here
		println("config changed:", v.Name)
	})
	// Add a listener if you want to monitor config changes
	return v
}

```

```go
package main



var configFile = flag.String("f", "etc/demoa.yaml", "the config file")

func main() {
	flag.Parse()

	// Load base config
	var b config.BaseConfig
	conf.MustLoad(*configFile, &b)

	// Subscribe to the consul config center
	var c config.Config
	c = config.SubscriberConsulConfig(b)

	ctx := svc.NewServiceContext(c)
	

}

```

## Configuration Parameters

```go
type Conf struct {
	Host       string        `json:",optional"`   // consul address
	Scheme     string        `json:",default=http"`  // consul address scheme
	PathPrefix string        `json:",optional"`  // 
	Datacenter string        `json:",optional"`  // datacenter
	Token      string        `json:",optional"`  // consul token
	TLSConfig  api.TLSConfig `json:"TLSConfig,optional"` // consul tls
	Key        string        `json:",optional"` // config center key name
	Type       string        `json:",default=yaml,options=yaml|hcl|json|xml"` // config type
}

```

## Changelog

See [CHANGELOG.md](./CHANGELOG.md)
