package demoA

import (
	"context"

	"example/example/internal/svc"
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

func (l *GDemoBLogic) Consume(ctx context.Context, message []byte) error {
	// todo: add your logic here and delete this line

	return nil
}
