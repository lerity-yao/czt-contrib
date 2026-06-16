package sdk

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gookit/color"
	"github.com/lerity-yao/czt-contrib/cztctl/util"
	"github.com/lerity-yao/czt-contrib/cztctl/util/pathx"
	"github.com/spf13/cobra"
)

var (
	// VarStringProto describes the proto file path.
	VarStringProto string
	// VarStringRepo describes the SDK repo full git URL.
	VarStringRepo string
	// VarStringRepoUser describes the repo auth username.
	VarStringRepoUser string
	// VarStringRepoToken describes the repo auth credential.
	VarStringRepoToken string
	// VarStringRemote describes the remote git repository of the template.
	VarStringRemote string
	// VarStringStyle describes the file naming style.
	VarStringStyle string
	// VarStringTag describes the SDK version tag.
	VarStringTag string
	// VarStringRepoBranch describes the SDK repo git branch name.
	VarStringRepoBranch string
	// VarStringGoProxy describes the Go module proxy URL.
	VarStringGoProxy string
)

// GoSdkCommand generates RPC client SDK from proto file.
func GoSdkCommand(_ *cobra.Command, _ []string) error {
	proto := VarStringProto
	repo := VarStringRepo
	repoUser := VarStringRepoUser
	repoToken := VarStringRepoToken
	remote := VarStringRemote
	branch := VarStringBranch
	style := VarStringStyle
	multiple := VarBoolMultiple
	tag := VarStringTag
	repoBranch := VarStringRepoBranch
	goproxy := VarStringGoProxy

	return DoGenSDK(proto, repo, repoUser, repoToken, remote, branch, style, multiple, tag, repoBranch, goproxy)
}

// DoGenSDK generates RPC client SDK from proto file.
func DoGenSDK(proto, repo, repoUser, repoToken, remote, branch, style string, multiple bool, tag, repoBranch, goproxy string) error {
	// Step 0: validate required parameters
	if len(proto) == 0 {
		return errors.New("missing --proto")
	}
	if len(repo) == 0 {
		return errors.New("missing --repo")
	}

	// Step 1: resolve proto absolute path and project root
	protoAbsPath, err := filepath.Abs(proto)
	if err != nil {
		return fmt.Errorf("failed to resolve proto path: %w", err)
	}

	// Find project root (go.mod directory) from proto file location
	_, projectRoot := pathx.FindGoModPath(filepath.Dir(protoAbsPath))
	if len(projectRoot) == 0 {
		return errors.New("failed to find project root (go.mod) from proto path")
	}

	// Parse --repo: remove scheme and .git suffix to get SDK_MODULE
	sdkModule, err := parseSDKModule(repo)
	if err != nil {
		return fmt.Errorf("failed to parse --repo: %w", err)
	}

	// Build AUTH_REPO: inject auth credentials into repo URL
	authRepo := util.BuildAuthURL(repo, repoUser, repoToken)

	sdkDir := sdkDirName

	// Step 2: check goctl availability
	if err := CheckGoctl(); err != nil {
		return err
	}

	// Step 3: git clone or init
	// Always start fresh — remove leftover _sdk/ from previous runs
	os.RemoveAll(sdkDir)
	fmt.Println(color.Green.Render(fmt.Sprintf("git clone %s %s", repo, sdkDir)))
	isUpdate, err := util.GitCloneOrInit(authRepo, sdkDir)
	if err != nil {
		return err
	}
	defer os.RemoveAll(sdkDir)

	// Step 4: update mode — clear client/ directory
	if isUpdate {
		fmt.Println(color.Green.Render("rm -rf _sdk/client/"))
		if err := util.ClearClientDir(sdkDir); err != nil {
			return err
		}
	}

	// Step 5: ensure go.mod exists (handles both init and empty-repo-clone)
	goModPath := filepath.Join(sdkDir, "go.mod")
	if _, statErr := os.Stat(goModPath); os.IsNotExist(statErr) {
		fmt.Println(color.Green.Render(fmt.Sprintf("go mod init %s", sdkModule)))
		if err := GoModInit(sdkDir, sdkModule); err != nil {
			return err
		}
	}

	// Step 6: copy proto files
	fmt.Println(color.Green.Render(fmt.Sprintf("cp proto -> %s/", sdkDir)))
	if err := CopyProtoRecursive(sdkDir, protoAbsPath, projectRoot); err != nil {
		return err
	}

	// Step 6.5: generate .kong.proto in _sdk/ (same dir as copied proto)
	relProtoPath, err := filepath.Rel(projectRoot, protoAbsPath)
	if err != nil {
		return fmt.Errorf("failed to compute relative proto path: %w", err)
	}
	sdkProtoAbsPath := filepath.Join(sdkDir, relProtoPath)
	kongPath, err := GenerateKongProto(sdkProtoAbsPath)
	if err != nil {
		return fmt.Errorf("failed to generate kong proto: %w", err)
	}
	fmt.Println(color.Green.Render(fmt.Sprintf("generated kong proto -> %s", kongPath)))

	// Step 7: relative proto path already computed in step 6.5

	// Step 8: run goctl rpc protoc
	goctlCmd := fmt.Sprintf("goctl rpc protoc %s --zrpc_out=. --go_out=./client/pb --go-grpc_out=./client/pb --style=%s", relProtoPath, style)
	if remote != "" {
		goctlCmd += fmt.Sprintf(" --remote %s", remote)
	}
	if branch != "" {
		goctlCmd += fmt.Sprintf(" --branch %s", branch)
	}
	if multiple {
		goctlCmd += " -m"
	}
	fmt.Println(color.Green.Render(goctlCmd))
	if err := GoctlRpcProtoc(sdkDir, relProtoPath, remote, branch, style, multiple); err != nil {
		return err
	}

	// Step 9: remove server-side code
	fmt.Println(color.Green.Render("rm -rf _sdk/internal/ _sdk/etc/ _sdk/*.go"))
	if err := CleanServerCode(sdkDir); err != nil {
		return err
	}

	// Step 10: go mod tidy
	modTidyCmd := "go mod tidy"
	if goproxy != "" {
		modTidyCmd = fmt.Sprintf("GOPROXY=%s go mod tidy", goproxy)
	}
	fmt.Println(color.Green.Render(modTidyCmd))
	if err := GoModTidy(sdkDir, goproxy); err != nil {
		return err
	}

	// Step 11: git commit + tag + push
	resolvedTag, err := util.GitCommitAndTag(sdkDir, authRepo, tag, repoBranch, isUpdate)
	if err != nil {
		return err
	}

	fmt.Println(color.Green.Render(fmt.Sprintf("Published %s", resolvedTag)))
	fmt.Println(color.Green.Render("Done."))
	return nil
}

// parseSDKModule parses the --repo URL and returns the SDK module path.
// It removes the scheme (http:// or https://) and .git suffix.
// Example: "https://gitlab.ddtz.com/backend/tax-invoice-rpc-sdk.git" → "gitlab.ddtz.com/backend/tax-invoice-rpc-sdk"
func parseSDKModule(repo string) (string, error) {
	u, err := url.Parse(repo)
	if err != nil {
		return "", fmt.Errorf("invalid repo URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("repo URL must start with http:// or https://, got: %s", repo)
	}
	host := u.Hostname() // 去掉端口
	if host == "" {
		return "", fmt.Errorf("repo URL must include host, got: %s", repo)
	}
	path := strings.TrimSuffix(strings.TrimRight(u.Path, "/"), ".git")
	if path == "" || path == "/" {
		return "", fmt.Errorf("repo URL must include repository path, got: %s", repo)
	}
	module := host + path
	return module, nil
}
