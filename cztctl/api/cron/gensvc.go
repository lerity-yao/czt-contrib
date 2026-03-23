package cron

import (
	_ "embed"
	"fmt"

	"github.com/lerity-yao/czt-contrib/cztctl/api/gogen"
	"github.com/lerity-yao/czt-contrib/cztctl/config"
	"github.com/lerity-yao/czt-contrib/cztctl/util/format"
	"github.com/lerity-yao/czt-contrib/cztctl/util/pathx"
	"github.com/lerity-yao/czt-contrib/cztctl/vars"
)

const contextFilename = "service_context"

//go:embed svc.tpl
var contextTemplate string

func genServiceContext(dir, rootPkg string, cfg *config.Config) error {
	filename, err := format.FileNamingFormat(cfg.NamingFormat, contextFilename)
	if err != nil {
		return err
	}

	importPackages := "\"" + pathx.JoinPackages(rootPkg, configDir) + "\"" + fmt.Sprintf("\n\"%s/czt-contrib/cron\"", vars.YaoxProjectOpenSourceURL)

	return gogen.GenFile(gogen.FileGenConfig{
		Dir:             dir,
		Subdir:          contextDir,
		Filename:        filename + ".go",
		TemplateName:    "contextTemplate",
		Category:        category,
		TemplateFile:    contextTemplateFile,
		BuiltinTemplate: contextTemplate,
		Data: map[string]string{
			"importPackages": importPackages,
			"config":         "config.Config",
		},
	})
}
