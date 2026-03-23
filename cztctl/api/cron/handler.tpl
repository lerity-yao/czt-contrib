// Code scaffolded by cztctl. Safe to edit.
// cztctl {{.version}}

package {{.PkgName}}

import (
    "context"
    {{if .HasRequest}}"encoding/json"{{end}}
	{{.ImportPackages}}
)

{{if .HasDoc}}{{.Doc}}{{end}}
func {{.HandlerName}}(svcCtx *svc.ServiceContext) cron.HandlerFunc {
	return func(ctx context.Context, t *cron.Task) error {
	   {{if .HasRequest}}var req types.{{.RequestType}}
        err := json.Unmarshal(t.Payload, &req)
        if err != nil {
        	return err
        }
        {{end}}l := {{.LogicName}}.New{{.LogicType}}(ctx, svcCtx)
        {{if .HasRequest}} return l.{{.function}}(&req)
        {{else}} return l.{{.function}}()
        {{end}}
    }
}

