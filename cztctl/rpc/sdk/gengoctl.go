package sdk

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/creack/pty"
)

// GoModInit runs go mod init in the SDK directory (init mode only).
func GoModInit(sdkDir, sdkModule string) error {
	cmd := exec.Command("go", "mod", "init", sdkModule)
	cmd.Dir = sdkDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to go mod init: %s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}

// GoctlRpcProtoc runs goctl rpc protoc to generate code in the SDK directory.
// Uses a pty for stderr to preserve TTY detection (git progress output),
// while filtering Usage/Flags/Examples help sections in real-time.
func GoctlRpcProtoc(sdkDir, protoPath, remote, branch, style string, multiple bool) error {
	args := []string{
		"rpc", "protoc", protoPath,
		"--zrpc_out=.",
		"--go_out=./client/pb",
		"--go-grpc_out=./client/pb",
		"--style=" + style,
	}
	if remote != "" {
		args = append(args, "--remote", remote)
	}
	if branch != "" {
		args = append(args, "--branch", branch)
	}
	if multiple {
		args = append(args, "-m")
	}
	cmd := exec.Command("goctl", args...)
	cmd.Dir = sdkDir
	cmd.Env = os.Environ()
	cmd.Stdout = os.Stdout

	// Create pty for stderr so child process sees a real terminal (git outputs progress)
	ptmx, tty, err := pty.Open()
	if err != nil {
		// Fallback: direct stderr if pty fails
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("goctl rpc protoc failed")
		}
		return nil
	}
	cmd.Stderr = tty

	var (
		errMsg string
		wg     sync.WaitGroup
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		errMsg = filterStderrFromPty(ptmx)
	}()

	cmdErr := cmd.Run()
	tty.Close() // close slave so reader gets EOF
	wg.Wait()   // wait for reader goroutine to finish
	ptmx.Close()

	if cmdErr != nil {
		if errMsg != "" {
			return fmt.Errorf("goctl rpc protoc failed: %s", errMsg)
		}
		return fmt.Errorf("goctl rpc protoc failed")
	}
	return nil
}

// filterStderrFromPty reads from the pty master in real-time,
// outputs everything except Usage/Flags/Examples sections,
// and captures the Error: line if present.
func filterStderrFromPty(r io.Reader) string {
	var (
		lineBuf    []byte
		suppressed bool
		errMsg     string
	)
	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			for _, b := range buf[:n] {
				if suppressed {
					continue
				}
				if b == '\n' {
					line := string(lineBuf)
					trimmed := strings.TrimSpace(line)
					if isUsageLine(trimmed) {
						suppressed = true
						lineBuf = nil
						continue
					}
					if strings.HasPrefix(trimmed, "Error:") {
						errMsg = strings.TrimSpace(strings.TrimPrefix(trimmed, "Error:"))
					}
					os.Stderr.Write(append(lineBuf, '\n'))
					lineBuf = nil
				} else if b == '\r' {
					// Progress update (e.g. git clone) - flush immediately
					os.Stderr.Write(append(lineBuf, '\r'))
					lineBuf = nil
				} else {
					lineBuf = append(lineBuf, b)
				}
			}
		}
		if err != nil {
			break
		}
	}
	// Flush remaining
	if len(lineBuf) > 0 && !suppressed {
		os.Stderr.Write(lineBuf)
	}
	return errMsg
}

func isUsageLine(trimmed string) bool {
	return strings.HasPrefix(trimmed, "Usage:") ||
		strings.HasPrefix(trimmed, "Flags:") ||
		strings.HasPrefix(trimmed, "Global Flags:") ||
		strings.HasPrefix(trimmed, "Examples:") ||
		strings.HasPrefix(trimmed, "Available Commands:")
}

// CleanServerCode removes server-side code from the SDK directory.
func CleanServerCode(sdkDir string) error {
	// Remove internal/ and etc/ directories
	if err := os.RemoveAll(filepath.Join(sdkDir, "internal")); err != nil {
		return fmt.Errorf("failed to remove internal/: %w", err)
	}
	if err := os.RemoveAll(filepath.Join(sdkDir, "etc")); err != nil {
		return fmt.Errorf("failed to remove etc/: %w", err)
	}

	// Remove all .go files in sdkDir root (non-recursive)
	entries, err := os.ReadDir(sdkDir)
	if err != nil {
		return fmt.Errorf("failed to read sdk directory: %w", err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.HasSuffix(entry.Name(), ".go") {
			if err := os.Remove(filepath.Join(sdkDir, entry.Name())); err != nil {
				return fmt.Errorf("failed to remove %s: %w", entry.Name(), err)
			}
		}
	}
	return nil
}

var protoImportRegex = regexp.MustCompile(`^import\s+"([^"]+)"\s*;`)

// CopyProtoRecursive copies proto files recursively to the SDK directory,
// preserving the path structure relative to projectRoot. It parses the main
// proto file's import statements, recursively collects all dependent proto
// files (skipping well-known types), and copies them so that goctl can resolve
// them using the same relative paths as in the original project.
func CopyProtoRecursive(sdkDir, protoAbsPath, projectRoot string) error {
	visited := make(map[string]bool)
	var files []string

	// Collect all proto files recursively
	if err := collectProtoFiles(protoAbsPath, projectRoot, visited, &files); err != nil {
		return err
	}

	// Copy all collected files preserving relative path
	for _, absPath := range files {
		relPath, err := filepath.Rel(projectRoot, absPath)
		if err != nil {
			return fmt.Errorf("failed to compute relative path for %s: %w", absPath, err)
		}
		destPath := filepath.Join(sdkDir, relPath)
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", destPath, err)
		}
		if err := copyFile(absPath, destPath); err != nil {
			return fmt.Errorf("failed to copy %s to %s: %w", absPath, destPath, err)
		}
	}
	return nil
}

func collectProtoFiles(protoAbsPath, projectRoot string, visited map[string]bool, files *[]string) error {
	if visited[protoAbsPath] {
		return nil
	}
	visited[protoAbsPath] = true
	*files = append(*files, protoAbsPath)

	imports, err := parseProtoImports(protoAbsPath)
	if err != nil {
		return err
	}

	for _, imp := range imports {
		if isWellKnownProto(imp) {
			continue
		}
		depAbsPath := filepath.Join(projectRoot, imp)
		if err := collectProtoFiles(depAbsPath, projectRoot, visited, files); err != nil {
			return err
		}
	}
	return nil
}

func parseProtoImports(protoPath string) ([]string, error) {
	f, err := os.Open(protoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open proto file %s: %w", protoPath, err)
	}
	defer f.Close()

	var imports []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		matches := protoImportRegex.FindStringSubmatch(line)
		if len(matches) == 2 {
			imports = append(imports, matches[1])
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan proto file %s: %w", protoPath, err)
	}
	return imports, nil
}

func isWellKnownProto(importPath string) bool {
	wellKnownPrefixes := []string{
		"google/protobuf/",
		"google/api/",
		"google/rpc/",
		"validate/",
	}
	for _, prefix := range wellKnownPrefixes {
		if strings.HasPrefix(importPath, prefix) {
			return true
		}
	}
	return false
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// GoModTidy runs go mod tidy in the SDK directory.
// If goproxy is non-empty, sets GOPROXY environment variable before execution.
func GoModTidy(sdkDir, goproxy string) error {
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = sdkDir
	cmd.Env = os.Environ()
	if goproxy != "" {
		cmd.Env = append(cmd.Env, "GOPROXY="+goproxy)
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to go mod tidy: %w", err)
	}
	return nil
}
