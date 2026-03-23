package cron

import (
	_ "embed"
	"fmt"

	"github.com/lerity-yao/czt-contrib/cztctl/api/gogen"
	"github.com/lerity-yao/czt-contrib/cztctl/api/spec"
	"github.com/lerity-yao/czt-contrib/cztctl/config"
	"github.com/lerity-yao/czt-contrib/cztctl/util/format"
	"github.com/lerity-yao/czt-contrib/cztctl/vars"
)

const configFile = "config"

//go:embed config.tpl
var configTemplate string

func genConfig(dir string, cfg *config.Config, api *spec.ApiSpec) error {
	filename, err := format.FileNamingFormat(cfg.NamingFormat, configFile)
	if err != nil {
		return err
	}

	authImportStr := fmt.Sprintf("\"%s/rest\"", vars.ProjectOpenSourceURL)
	authImportStr = authImportStr + fmt.Sprintf("\n\"%s/czt-contrib/cron\"", vars.YaoxProjectOpenSourceURL)
	configName := fmt.Sprintf("CronConf cron.ServerConfig\n")

	return gogen.GenFile(gogen.FileGenConfig{
		Dir:             dir,
		Subdir:          configDir,
		Filename:        filename + ".go",
		TemplateName:    "configTemplate",
		Category:        category,
		TemplateFile:    configTemplateFile,
		BuiltinTemplate: configTemplate,
		Data: map[string]string{
			"authImport": authImportStr,
			"cronConfig": configName,
		},
	})
}
