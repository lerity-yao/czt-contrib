package sdk

const (
	// DefaultRepoUser is the default repo auth username for push.
	DefaultRepoUser = "cztctl-bot"
	// DefaultRepoToken is the default repo auth credential (password/token/key) for push.
	DefaultRepoToken = "glpat-Cvi30LCsPVteDGdPWxmER286MQp1OjUH.01.0w0iuu5yw"
	// DefaultRemote is the default remote template repository (empty means use goctl default).
	DefaultRemote = ""
	// DefaultRepoBranch is the default git branch name for SDK repo.
	DefaultRepoBranch = "main"

	sdkDirName = "_sdk"
)

var VarStringBranch string
var VarBoolMultiple bool
