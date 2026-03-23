package api

import (
	"github.com/lerity-yao/czt-contrib/cztctl/api/cron"
	"github.com/lerity-yao/czt-contrib/cztctl/api/rabbitmq"
	"github.com/lerity-yao/czt-contrib/cztctl/api/swagger"
	"github.com/lerity-yao/czt-contrib/cztctl/config"
	"github.com/lerity-yao/czt-contrib/cztctl/internal/cobrax"
)

var (
	// Cmd describes an api command.
	Cmd           = cobrax.NewCommand("api")
	swaggerCmd    = cobrax.NewCommand("swagger", cobrax.WithRunE(swagger.Command))
	goRabbitmqCmd = cobrax.NewCommand("rabbitmq", cobrax.WithRunE(rabbitmq.GoRabbitmqCommand))
	goCronCmd     = cobrax.NewCommand("cron", cobrax.WithRunE(cron.GoCronCommand))
)

func init() {
	var (
		swaggerCmdFlags    = swaggerCmd.Flags()
		goRabbitmqCmdFlags = goRabbitmqCmd.Flags()
		goCronCmdFlags     = goCronCmd.Flags()
	)

	goRabbitmqCmdFlags.StringVar(&rabbitmq.VarStringDir, "dir")
	goRabbitmqCmdFlags.StringVar(&rabbitmq.VarStringAPI, "api")
	goRabbitmqCmdFlags.StringVar(&rabbitmq.VarStringHome, "home")
	goRabbitmqCmdFlags.StringVar(&rabbitmq.VarStringRemote, "remote")
	goRabbitmqCmdFlags.StringVar(&rabbitmq.VarStringBranch, "branch")
	goRabbitmqCmdFlags.BoolVar(&rabbitmq.VarBoolWithTest, "test")
	goRabbitmqCmdFlags.BoolVar(&rabbitmq.VarBoolTypeGroup, "type-group")
	goRabbitmqCmdFlags.StringVarWithDefaultValue(&rabbitmq.VarStringStyle, "style", config.DefaultFormat)

	goCronCmdFlags.StringVar(&cron.VarStringDir, "dir")
	goCronCmdFlags.StringVar(&cron.VarStringAPI, "api")
	goCronCmdFlags.StringVar(&cron.VarStringHome, "home")
	goCronCmdFlags.StringVar(&cron.VarStringRemote, "remote")
	goCronCmdFlags.StringVar(&cron.VarStringBranch, "branch")
	goCronCmdFlags.BoolVar(&cron.VarBoolWithTest, "test")
	goCronCmdFlags.BoolVar(&cron.VarBoolTypeGroup, "type-group")
	goCronCmdFlags.StringVarWithDefaultValue(&cron.VarStringStyle, "style", config.DefaultFormat)

	swaggerCmdFlags.StringVar(&swagger.VarStringAPI, "api")
	swaggerCmdFlags.StringVar(&swagger.VarStringDir, "dir")
	swaggerCmdFlags.StringVar(&swagger.VarStringFilename, "filename")
	swaggerCmdFlags.BoolVar(&swagger.VarBoolYaml, "yaml")

	// Add sub-commands
	Cmd.AddCommand(swaggerCmd, goRabbitmqCmd, goCronCmd)
}
