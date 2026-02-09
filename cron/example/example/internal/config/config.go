package config

import (
	"github.com/lerity-yao/czt-contrib/cron"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	CronConf cron.ServerConfig
}
