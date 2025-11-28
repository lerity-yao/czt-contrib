# configcenter/consul

`go-zero` 配套的 `consul` 注册中心

## 使用

在配置文件中添加以下配置内容
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

// BaseConfig 基础配置
// BaseConfig 目前只放配置中心配置，允许项目在只配置配置中心配置的时候，顺利连接配置中心，获取其他配置文件
// BaseConfig 如有比 配置中心的配置优先级别高的，可放此配置中，否则禁止进入 BaseConfig，应该进入 Config
type BaseConfig struct {
	ConfigCenterConsul configCenterConsul.ConsulConf // 配置中心配置
}

// Config 项目配置
type Config struct {
	rest.RestConf
}

// SubscriberConsulConfig 订阅consul配置中心
// SubscriberConsulConfig 支持监控配置变化，但是不支持热重载配置
// SubscriberConsulConfig k8s不同于其他，配置变化，需要重启pod，所以不支持重载
func SubscriberConsulConfig(b BaseConfig) Config {

	ss := configCenterConsul.MustNewConsulSubscriber(b.ConfigCenterConsul)
	// 创建 configurator
	cc := configurator.MustNewConfigCenter[Config](configurator.Config{
		Type: "yaml", // 配置值类型：json,yaml,toml
	}, ss)

	// 获取配置
	// 注意: 配置如果发生变更，调用的结果永远获取到最新的配置
	v, err := cc.GetConfig()
	if err != nil {
		panic(err)
	}
	cc.AddListener(func() {
		v, err := cc.GetConfig()
		if err != nil {
			panic(err)
		}
		//这个地方要写 触发配置变化后 需要处理的操作
		println("config changed:", v.Name)
	})
	// 如果想监听配置变化，可以添加 listener
	return v
}

```

```go
package main



var configFile = flag.String("f", "etc/demoa.yaml", "the config file")

func main() {
	flag.Parse()

	// 加载基础配置
	var b config.BaseConfig
	conf.MustLoad(*configFile, &b)

	// 订阅 consul配置中心
	var c config.Config
	c = config.SubscriberConsulConfig(b)

	ctx := svc.NewServiceContext(c)
	

}

```

## 配置参数

```go
type Conf struct {
	Host       string        `json:",optional"`   // consul 地址
	Scheme     string        `json:",default=http"`  // consul地址scheme
	PathPrefix string        `json:",optional"`  // 
	Datacenter string        `json:",optional"`  // 数据中心
	Token      string        `json:",optional"`  // consul token
	TLSConfig  api.TLSConfig `json:"TLSConfig,optional"` // consul tls
	Key        string        `json:",optional"` // 配置中心key名称
	Type       string        `json:",default=yaml,options=yaml|hcl|json|xml"` // 配置类型
}
