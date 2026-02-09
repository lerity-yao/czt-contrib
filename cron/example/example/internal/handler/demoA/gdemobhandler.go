// Code scaffolded by cztctl. Safe to edit.
// cztctl 1.9.4.2

package demoA

import (
	"context"

	"example/example/internal/logic/demoA"
	"example/example/internal/svc"
	"github.com/lerity-yao/czt-contrib/cron"
)

// cron: */1 * * * *
func GDemoBHandler(svcCtx *svc.ServiceContext) cron.HandlerFunc {
	return func(ctx context.Context, t *cron.Task) error {
		l := demoA.NewGDemoBLogic(ctx, svcCtx)
		return l.GDemoB()

	}
}
