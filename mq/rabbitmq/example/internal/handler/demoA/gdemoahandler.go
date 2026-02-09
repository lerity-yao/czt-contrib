package demoA

import (
	"context"
	"example/example/internal/logic/demoA"
	"example/example/internal/svc"
	"github.com/lerity-yao/czt-contrib/mq/rabbitmq"
	"github.com/zeromicro/go-zero/core/service"
)

func GDemoAHandler(ctx context.Context, svcCtx *svc.ServiceContext) service.Service {
	return rabbitmq.MustNewListener(ctx, svcCtx.Config.GDemoARabbitmqConf, demoA.NewGDemoALogic(ctx, svcCtx))
}
