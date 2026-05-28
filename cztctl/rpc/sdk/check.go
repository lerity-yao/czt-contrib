package sdk

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/gookit/color"
	"github.com/lerity-yao/czt-contrib/cztctl/internal/version"
)

// CheckGoctl checks if goctl is installed and version is compatible.
func CheckGoctl() error {
	_, err := exec.LookPath("goctl")
	if err != nil {
		return fmt.Errorf("goctl is not installed, please run: go install github.com/zeromicro/go-zero/tools/goctl@latest")
	}

	output, err := exec.Command("goctl", "--version").CombinedOutput()
	if err != nil {
		fmt.Printf("%s\n", color.Yellow.Render(
			fmt.Sprintf("WARNING: failed to get goctl version: %v", err),
		))
		return nil
	}

	goctlVer, err := parseGoctlVersion(string(output))
	if err != nil {
		fmt.Printf("%s\n", color.Yellow.Render(
			fmt.Sprintf("WARNING: %v", err),
		))
		return nil
	}

	expectedVer := version.GetGoZeroVersion()
	expectedVer = strings.TrimPrefix(expectedVer, "v")

	if compareVersions(goctlVer, expectedVer) < 0 {
		fmt.Printf("%s\n", color.Yellow.Render(
			fmt.Sprintf("WARNING: goctl version %s is older than expected %s, consider upgrading via: go install github.com/zeromicro/go-zero/tools/goctl@latest", goctlVer, expectedVer),
		))
	}

	return nil
}

// parseGoctlVersion parses the version number from goctl --version output.
// Expected format: "goctl version 1.7.3 linux/amd64" or similar.
func parseGoctlVersion(output string) (string, error) {
	output = strings.TrimSpace(output)
	fields := strings.Fields(output)
	for _, field := range fields {
		parts := strings.Split(field, ".")
		if len(parts) == 3 {
			if _, err := strconv.Atoi(parts[0]); err == nil {
				if _, err := strconv.Atoi(parts[1]); err == nil {
					if _, err := strconv.Atoi(parts[2]); err == nil {
						return field, nil
					}
				}
			}
		}
	}
	return "", fmt.Errorf("cannot parse version from output: %s", output)
}

// compareVersions compares two semver strings (without "v" prefix).
// Returns -1 if v1 < v2, 0 if v1 == v2, 1 if v1 > v2.
func compareVersions(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")

	for i := 0; i < 3; i++ {
		n1, _ := strconv.Atoi(parts1[i])
		n2, _ := strconv.Atoi(parts2[i])
		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}
	return 0
}
