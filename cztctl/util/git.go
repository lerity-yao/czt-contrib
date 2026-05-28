package util

import (
	"bytes"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gookit/color"
	"github.com/lerity-yao/czt-contrib/cztctl/internal/version"
	"github.com/lerity-yao/czt-contrib/cztctl/util/env"
	"github.com/lerity-yao/czt-contrib/cztctl/util/pathx"
)

// runGit executes a git command in the given working directory and returns
// an error containing stderr output on failure.
func runGit(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg != "" {
			return fmt.Errorf("git %s failed: %s: %w", strings.Join(args, " "), msg, err)
		}
		return fmt.Errorf("git %s failed: %w", strings.Join(args, " "), err)
	}
	return nil
}

// runGitLive executes a git command with stdout/stderr directly connected to os,
// so that progress output (e.g. git push) is visible in real-time.
func runGitLive(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s failed: %w", strings.Join(args, " "), err)
	}
	return nil
}

// BuildAuthURL builds the authenticated repo URL by injecting credentials into the repo URL.
// Example: "https://gitlab.ddtz.com/x.git" → "https://{user}:{token}@gitlab.ddtz.com/x.git"
func BuildAuthURL(repoURL, repoUser, repoToken string) string {
	u, err := url.Parse(repoURL)
	if err != nil {
		// Fallback: return original URL if parsing fails
		return repoURL
	}
	if repoUser != "" || repoToken != "" {
		u.User = url.UserPassword(repoUser, repoToken)
	}
	return u.String()
}

// GitCloneOrInit tries to git clone the SDK repo, returns isUpdate=true on success.
// On success, sdkDir is guaranteed to exist:
//   - clone success → git creates sdkDir with repo content → isUpdate=true
//   - clone fails (repo doesn't exist) → MkdirAll creates empty dir → isUpdate=false
func GitCloneOrInit(authRepo, sdkDir string) (isUpdate bool, err error) {
	// Try git clone: if success, git creates sdkDir with repo content
	if cloneErr := runGit("", "clone", authRepo, sdkDir); cloneErr == nil {
		return true, nil
	}

	// Clone failed — treat it as initialization mode (repo doesn't exist yet).
	if err := os.MkdirAll(sdkDir, 0755); err != nil {
		return false, fmt.Errorf("failed to create %s directory: %w", sdkDir, err)
	}
	return false, nil
}

// GitInitAndRemote initializes a git repo and adds the remote with auth URL.
func GitInitAndRemote(sdkDir, authRepo string) error {
	if err := runGit(sdkDir, "init"); err != nil {
		return err
	}
	if err := runGit(sdkDir, "remote", "add", "origin", authRepo); err != nil {
		return err
	}
	return nil
}

// GitCommitAndTag commits all changes, creates and pushes the tag.
// It first resolves the final tag via version.ResolveTag based on the latest
// tag in sdkDir, then performs git operations.
// For init mode (!isUpdate): git init + remote add + commit + tag + push -u origin {repoBranch} + push tag
// For update mode (isUpdate):  commit + tag + push origin {repoBranch} + push tag
func GitCommitAndTag(sdkDir, authRepo, tag, repoBranch string, isUpdate bool) (resolvedTag string, err error) {
	var latestTag string
	if isUpdate {
		latestTag, err = version.GetLatestTag(sdkDir)
		if err != nil {
			return "", fmt.Errorf("failed to get latest tag: %w", err)
		}
	}

	resolvedTag, err = version.ResolveTag(tag, latestTag)
	if err != nil {
		return "", fmt.Errorf("failed to resolve tag: %w", err)
	}

	commitMsg := fmt.Sprintf("release %s", resolvedTag)

	if !isUpdate {
		fmt.Println(color.Green.Render("git init"))
		if err := runGit(sdkDir, "init"); err != nil {
			return "", err
		}
		if err := runGit(sdkDir, "remote", "add", "origin", authRepo); err != nil {
			return "", err
		}
		fmt.Println(color.Green.Render("git add -A"))
		if err := runGit(sdkDir, "add", "-A"); err != nil {
			return "", err
		}
		fmt.Println(color.Green.Render(fmt.Sprintf("git commit -m \"%s\"", commitMsg)))
		if err := runGit(sdkDir, "commit", "-m", commitMsg); err != nil {
			return "", err
		}
		// Ensure local branch matches repoBranch (git init may default to 'main')
		if err := runGit(sdkDir, "branch", "-M", repoBranch); err != nil {
			return "", err
		}
		fmt.Println(color.Green.Render(fmt.Sprintf("git tag %s", resolvedTag)))
		if err := runGit(sdkDir, "tag", resolvedTag); err != nil {
			return "", err
		}
		fmt.Println(color.Green.Render(fmt.Sprintf("git push -u origin %s", repoBranch)))
		if err := runGitLive(sdkDir, "push", "-u", "origin", repoBranch); err != nil {
			return "", err
		}
		fmt.Println(color.Green.Render(fmt.Sprintf("git push origin %s", resolvedTag)))
		if err := runGitLive(sdkDir, "push", "origin", resolvedTag); err != nil {
			return "", err
		}
		return resolvedTag, nil
	}

	// Update mode
	fmt.Println(color.Green.Render("git add -A"))
	if err := runGit(sdkDir, "add", "-A"); err != nil {
		return "", err
	}
	// Check if there are changes to commit
	if !hasChanges(sdkDir) {
		return "", fmt.Errorf("no changes to commit, SDK code is up-to-date")
	}
	fmt.Println(color.Green.Render(fmt.Sprintf("git commit -m \"%s\"", commitMsg)))
	if err := runGit(sdkDir, "commit", "-m", commitMsg); err != nil {
		return "", err
	}
	// Ensure local branch matches repoBranch (handles empty repo clone where default may be 'main')
	if err := runGit(sdkDir, "branch", "-M", repoBranch); err != nil {
		return "", err
	}
	fmt.Println(color.Green.Render(fmt.Sprintf("git tag %s", resolvedTag)))
	if err := runGit(sdkDir, "tag", resolvedTag); err != nil {
		return "", err
	}
	fmt.Println(color.Green.Render(fmt.Sprintf("git push origin %s", repoBranch)))
	if err := runGitLive(sdkDir, "push", "origin", repoBranch); err != nil {
		return "", err
	}
	fmt.Println(color.Green.Render(fmt.Sprintf("git push origin %s", resolvedTag)))
	if err := runGitLive(sdkDir, "push", "origin", resolvedTag); err != nil {
		return "", err
	}
	return resolvedTag, nil
}

// ClearClientDir removes the client/ directory in sdkDir if it exists.
func ClearClientDir(sdkDir string) error {
	clientDir := filepath.Join(sdkDir, "client")
	if err := os.RemoveAll(clientDir); err != nil {
		return fmt.Errorf("failed to remove %s: %w", clientDir, err)
	}
	return nil
}

// hasChanges returns true if there are staged changes in the git working tree.
func hasChanges(dir string) bool {
	cmd := exec.Command("git", "diff", "--cached", "--quiet")
	if dir != "" {
		cmd.Dir = dir
	}
	// exit 0 = no diff (no changes), exit 1 = has changes
	return cmd.Run() != nil
}

// CloneIntoGitHome clones a remote git repository into the cztctl git home directory.
func CloneIntoGitHome(url, branch string) (dir string, err error) {
	gitHome, err := pathx.GetGitHome()
	if err != nil {
		return "", err
	}
	os.RemoveAll(gitHome)
	ext := filepath.Ext(url)
	repo := strings.TrimSuffix(filepath.Base(url), ext)
	dir = filepath.Join(gitHome, repo)
	if pathx.FileExists(dir) {
		os.RemoveAll(dir)
	}
	path, err := env.LookPath("git")
	if err != nil {
		return "", err
	}
	if !env.CanExec() {
		return "", fmt.Errorf("os %q can not call 'exec' command", runtime.GOOS)
	}
	args := []string{"clone"}
	if len(branch) > 0 {
		args = append(args, "-b", branch)
	}
	args = append(args, url, dir)
	cmd := exec.Command(path, args...)
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	return
}
