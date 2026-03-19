package demoA

import (
	"context"
	"github.com/lerity-yao/czt-contrib/mq/rabbitmq"
	"github.com/lerity-yao/czt-contrib/mq/rabbitmq/example/internal/logic/demoA"
	"github.com/lerity-yao/czt-contrib/mq/rabbitmq/example/internal/svc"
	"github.com/zeromicro/go-zero/core/service"
)

func GDemoAHandler(svcCtx *svc.ServiceContext) service.Service {
	handler := func(ctx context.Context, message []byte) error {
		l := demoA.NewGDemoALogic(ctx, svcCtx)
		return l.GDemoA(message)
	}
	return rabbitmq.MustNewListener(svcCtx.Config.GDemoARabbitmqConf, rabbitmq.HandlerFunc(handler))
}
