package {{.PkgName}}

import (
    "context"
	{{.ImportPackages}}
)

{{if .HasDoc}}{{.Doc}}{{end}}
func {{.HandlerName}}(svcCtx *svc.ServiceContext) service.Service {
	handler := func(ctx context.Context, req []byte) error {
		l := {{.LogicName}}.New{{.LogicType}}(ctx, svcCtx)
		return l.{{.function}}(req)
	}
    return rabbitmq.MustNewListener(svcCtx.Config.{{.RabbitmqConfName}}, rabbitmq.HandlerFunc(handler))
}

