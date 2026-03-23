package svc

import (
    {{.importPackages}}
)

type ServiceContext struct {
	Config {{.config}}
	CronServer   cron.Server
}

func NewServiceContext(c {{.config}}) *ServiceContext {
	c.CronConf.Namespace = c.Name
	cronServer := cron.MustNewServer(c.CronConf, cron.WithServerLogger(&cron.AsynqLogger{}))

	return &ServiceContext{
		Config: c,
		CronServer:   cronServer,
	}
}
