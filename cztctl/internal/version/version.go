package version

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"runtime/debug"
	"strings"
)

const (
	BuildVersion = "1.10.6"
)

// GetGoctlVersion returns BuildVersion
func GetGoctlVersion() string {
	return BuildVersion
}

// GetGoZeroVersion reads the go-zero dependency version from build info.
func GetGoZeroVersion() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "unknown"
	}
	for _, dep := range info.Deps {
		if dep.Path == "github.com/zeromicro/go-zero" {
			if dep.Replace != nil {
				return dep.Replace.Version
			}
			return dep.Version
		}
	}
	return "unknown"
}

// SemVer represents a semantic version.
type SemVer struct {
	Major int
	Minor int
	Patch int
}

// ParseSemVer parses a version tag string into SemVer.
func ParseSemVer(tag string) (SemVer, error) {
	var major, minor, patch int
	n, err := fmt.Sscanf(tag, "v%d.%d.%d", &major, &minor, &patch)
	if err != nil || n != 3 {
		return SemVer{}, fmt.Errorf("invalid semver tag %q: must match v{major}.{minor}.{patch}", tag)
	}
	// Ensure no trailing characters
	expected := fmt.Sprintf("v%d.%d.%d", major, minor, patch)
	if tag != expected {
		return SemVer{}, fmt.Errorf("invalid semver tag %q: must match v{major}.{minor}.{patch}", tag)
	}
	return SemVer{Major: major, Minor: minor, Patch: patch}, nil
}

// ValidateSemVer validates a SemVer against the SDK tag rules.
func ValidateSemVer(sv SemVer) error {
	if sv.Major != 0 && sv.Major != 1 {
		return fmt.Errorf("major version must be 0 or 1, got %d", sv.Major)
	}
	if sv.Major == 0 && sv.Minor > 999 {
		return fmt.Errorf("when major=0, minor must be <= 999, got %d", sv.Minor)
	}
	return nil
}

// LessThan returns true if sv is less than other.
func (sv SemVer) LessThan(other SemVer) bool {
	if sv.Major != other.Major {
		return sv.Major < other.Major
	}
	if sv.Minor != other.Minor {
		return sv.Minor < other.Minor
	}
	return sv.Patch < other.Patch
}

// BumpMinor returns a new SemVer with minor incremented.
func (sv SemVer) BumpMinor() SemVer {
	return SemVer{Major: sv.Major, Minor: sv.Minor + 1, Patch: 0}
}

// BumpPatch returns a new SemVer with patch incremented.
// When patch exceeds 99, it resets to 0 and increments minor.
func (sv SemVer) BumpPatch() SemVer {
	if sv.Patch+1 > 99 {
		return SemVer{Major: sv.Major, Minor: sv.Minor + 1, Patch: 0}
	}
	return SemVer{Major: sv.Major, Minor: sv.Minor, Patch: sv.Patch + 1}
}

// String returns the string representation of SemVer.
func (sv SemVer) String() string {
	return fmt.Sprintf("v%d.%d.%d", sv.Major, sv.Minor, sv.Patch)
}

// ResolveTag resolves the final tag based on user input and latest tag.
func ResolveTag(userTag string, latestTag string) (string, error) {
	if latestTag == "" {
		// Initialization mode
		if userTag == "" {
			return "v1.0.0", nil
		}
		sv, err := ParseSemVer(userTag)
		if err != nil {
			return "", err
		}
		if err := ValidateSemVer(sv); err != nil {
			return "", err
		}
		return userTag, nil
	}

	// Update mode
	if userTag == "" {
		latestSV, err := ParseSemVer(latestTag)
		if err != nil {
			return "", fmt.Errorf("failed to parse latest tag %q: %w", latestTag, err)
		}
		newSV := latestSV.BumpPatch()
		return newSV.String(), nil
	}

	// User specified tag in update mode
	userSV, err := ParseSemVer(userTag)
	if err != nil {
		return "", err
	}
	if err := ValidateSemVer(userSV); err != nil {
		return "", err
	}
	latestSV, err := ParseSemVer(latestTag)
	if err != nil {
		return "", fmt.Errorf("failed to parse latest tag %q: %w", latestTag, err)
	}
	if !latestSV.LessThan(userSV) {
		return "", fmt.Errorf("tag %s must be greater than latest tag %s (downgrade not allowed)", userTag, latestTag)
	}
	return userTag, nil
}

// GetLatestTag returns the latest tag in the SDK git repository.
func GetLatestTag(sdkDir string) (string, error) {
	cmd := exec.Command("git", "tag", "--sort=-v:refname")
	cmd.Dir = sdkDir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git tag failed: %s: %w", strings.TrimSpace(stderr.String()), err)
	}
	scanner := bufio.NewScanner(&stdout)
	if scanner.Scan() {
		return strings.TrimSpace(scanner.Text()), nil
	}
	// No tags found — initialization mode
	return "", nil
}
