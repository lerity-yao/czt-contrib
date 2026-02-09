package config

import (
	"github.com/lerity-yao/czt-contrib/mq/rabbitmq"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	GDemoARabbitmqConf rabbitmq.RabbitListenerConf
	GDemoBRabbitmqConf rabbitmq.RabbitListenerConf
}
