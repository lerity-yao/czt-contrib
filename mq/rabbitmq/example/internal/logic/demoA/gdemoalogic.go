package demoA

import (
	"context"

	"example/example/internal/svc"
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

func (l *GDemoALogic) Consume(ctx context.Context, message []byte) error {
	// todo: add your logic here and delete this line

	return nil
}
