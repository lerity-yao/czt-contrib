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
