package env

import "github.com/lerity-yao/czt-contrib/cztctl/internal/cobrax"

var (
	sliceVarWriteValue []string

	// Cmd describes an env command.
	Cmd = cobrax.NewCommand("env", cobrax.WithRunE(write))
)

func init() {
	Cmd.Flags().StringSliceVarP(&sliceVarWriteValue, "write", "w")
}
