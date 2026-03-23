package apiutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// MaybeCreateFile creates a file if it does not exist.
// Returns the file handle, whether the file was created, and any error.
func MaybeCreateFile(dir, subdir, filename string) (*os.File, bool, error) {
	fpath := filepath.Join(dir, subdir, filename)
	fdir := filepath.Dir(fpath)
	if err := os.MkdirAll(fdir, os.ModePerm); err != nil {
		return nil, false, err
	}
	if _, err := os.Stat(fpath); err == nil {
		fmt.Printf("%s exists, ignored generation\n", fpath)
		return nil, false, nil
	}
	f, err := os.Create(fpath)
	return f, err == nil, err
}

// Copy copies src file to dst file.
func Copy(src, dst string) (int64, error) {
	sourceFile, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer sourceFile.Close()

	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, os.ModePerm); err != nil {
		return 0, err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destFile.Close()

	return io.Copy(destFile, sourceFile)
}

// WrapErr wraps an error with additional message.
func WrapErr(err error, message string) error {
	return fmt.Errorf("%s: %w", message, err)
}
