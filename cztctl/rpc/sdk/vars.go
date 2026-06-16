package sdk

const (
	// DefaultRepoUser is the default repo auth username for push.
	DefaultRepoUser = "cztctl-bot"
	// DefaultRepoToken is the default repo auth credential (password/token/key) for push.
	DefaultRepoToken = "glpat-rinPIxpXXi0v7pAUKAKQTG86MQp1OjEH.01.0w13qos5x"
	// DefaultRemote is the default remote template repository (empty means use goctl default).
	DefaultRemote = ""
	// DefaultRepoBranch is the default git branch name for SDK repo.
	DefaultRepoBranch = "main"

	sdkDirName = "_sdk"
)

var VarStringBranch string
var VarBoolMultiple bool
