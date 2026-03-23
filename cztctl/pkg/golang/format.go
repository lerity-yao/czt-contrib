package golang

import (
	"errors"
	"go/format"
	"path/filepath"

	"github.com/lerity-yao/czt-contrib/cztctl/util/pathx"
)

// FormatCode formats Go source code.
func FormatCode(code string) string {
	ret, err := format.Source([]byte(code))
	if err != nil {
		return code
	}
	return string(ret)
}

// GetParentPackage derives the Go package path for the given directory
// by finding go.mod and computing the relative path.
func GetParentPackage(dir string) (string, string, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", "", err
	}

	modPath, modDir := pathx.FindGoModPath(abs)
	if modPath == "" {
		return "", "", errors.New("go.mod not found")
	}

	rel, err := filepath.Rel(modDir, abs)
	if err != nil {
		return "", "", err
	}

	if rel == "." {
		return modPath, modPath, nil
	}

	return filepath.ToSlash(filepath.Join(modPath, rel)), modPath, nil
}
