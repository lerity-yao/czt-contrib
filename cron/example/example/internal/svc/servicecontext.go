package svc

import (
	"example/example/internal/config"
	"github.com/lerity-yao/czt-contrib/cron"
)

type ServiceContext struct {
	Config     config.Config
	CronServer cron.Server
}

func NewServiceContext(c config.Config) *ServiceContext {
	c.CronConf.Namespace = c.Name
	cronServer := cron.MustNewServer(c.CronConf, cron.WithServerLogger(&cron.AsynqLogger{}))

	return &ServiceContext{
		Config:     c,
		CronServer: cronServer,
	}
}
