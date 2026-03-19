package demoA

import (
	"context"

	"github.com/lerity-yao/czt-contrib/mq/rabbitmq/example/internal/svc"
	"github.com/zeromicro/go-zero/core/logx"
)

type GDemoBLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGDemoBLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GDemoBLogic {
	return &GDemoBLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GDemoBLogic) GDemoB(message []byte) error {
	// todo: add your logic here and delete this line

	return nil
}
