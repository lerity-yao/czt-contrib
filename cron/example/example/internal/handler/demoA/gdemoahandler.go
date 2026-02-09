// Code scaffolded by cztctl. Safe to edit.
// cztctl 1.9.4.2

package demoA

import (
	"context"
	"encoding/json"
	"example/example/internal/logic/demoA"
	"example/example/internal/svc"
	"example/example/internal/types"
	"github.com/lerity-yao/czt-contrib/cron"
)

func GDemoAHandler(svcCtx *svc.ServiceContext) cron.HandlerFunc {
	return func(ctx context.Context, t *cron.Task) error {
		var req types.Name
		err := json.Unmarshal(t.Payload, &req)
		if err != nil {
			return err
		}
		l := demoA.NewGDemoALogic(ctx, svcCtx)
		return l.GDemoA(&req)

	}
}
