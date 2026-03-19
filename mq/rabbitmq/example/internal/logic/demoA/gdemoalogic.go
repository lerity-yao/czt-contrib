package demoA

import (
	"context"

	"github.com/lerity-yao/czt-contrib/mq/rabbitmq/example/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type GDemoALogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGDemoALogic(ctx context.Context, svcCtx *svc.ServiceContext) *GDemoALogic {
	return &GDemoALogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GDemoALogic) GDemoA(message []byte) error {
	// todo: add your logic here and delete this line

	return nil
}
