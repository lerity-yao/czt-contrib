package rpc

import (
	"github.com/lerity-yao/czt-contrib/cztctl/config"
	"github.com/lerity-yao/czt-contrib/cztctl/internal/cobrax"
	"github.com/lerity-yao/czt-contrib/cztctl/rpc/sdk"
)

var (
	// Cmd describes an rpc command.
	Cmd    = cobrax.NewCommand("rpc")
	sdkCmd = cobrax.NewCommand("sdk", cobrax.WithRunE(sdk.GoSdkCommand))
)

func init() {
	var (
		sdkCmdFlags = sdkCmd.Flags()
	)

	sdkCmdFlags.StringVar(&sdk.VarStringProto, "proto")
	sdkCmdFlags.StringVar(&sdk.VarStringRepo, "repo")
	sdkCmdFlags.StringVarWithDefaultValue(&sdk.VarStringRepoUser, "repo-user", sdk.DefaultRepoUser)
	sdkCmdFlags.StringVarWithDefaultValue(&sdk.VarStringRepoToken, "repo-token", sdk.DefaultRepoToken)
	sdkCmdFlags.StringVarWithDefaultValue(&sdk.VarStringRemote, "remote", sdk.DefaultRemote)
	sdkCmdFlags.StringVar(&sdk.VarStringBranch, "branch")
	sdkCmdFlags.StringVarWithDefaultValue(&sdk.VarStringStyle, "style", config.DefaultFormat)
	sdkCmdFlags.BoolVarP(&sdk.VarBoolMultiple, "multiple", "m")
	sdkCmdFlags.StringVar(&sdk.VarStringTag, "tag")
	sdkCmdFlags.StringVarWithDefaultValue(&sdk.VarStringRepoBranch, "repo-branch", sdk.DefaultRepoBranch)
	sdkCmdFlags.StringVar(&sdk.VarStringGoProxy, "goproxy")

	// Add sub-commands
	Cmd.AddCommand(sdkCmd)
}
