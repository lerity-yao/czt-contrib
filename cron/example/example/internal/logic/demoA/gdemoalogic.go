// Code scaffolded by cztctl. Safe to edit.
// cztctl 1.9.4.2

package demoA

import (
	"context"

	"example/example/internal/svc"
	"example/example/internal/types"

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

func (l *GDemoALogic) GDemoA(req *types.Name) error {
	// todo: add your logic here and delete this line

	return nil
}
