package svc

import (
	"github.com/lerity-yao/czt-contrib/mq/rabbitmq/example/internal/config"
)

type ServiceContext struct {
	Config config.Config
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config: c,
	}
}
