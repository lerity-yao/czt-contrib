package sdk

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// serviceStartRe matches a service block opening line.
	// Service name is a proto identifier (\w+); using ".+" would greedily
	// swallow the opening brace "{".
	serviceStartRe = regexp.MustCompile(`^service\s+(\w+)\s*\{`)
	// rpcMethodRe matches an rpc method declaration line by the "returns"
	// keyword (a real rpc method always has returns), capturing leading
	// indent and the method name.
	rpcMethodRe = regexp.MustCompile(`^(\s*)rpc\s+(\w+).*\breturns\b`)
)

// GenerateKongProto reads a proto file and generates a .kong.proto variant
// with google.api.http annotations (POST /{ServiceName}/{RpcMethodName} body: "*").
func GenerateKongProto(protoPath string) (string, error) {
	data, err := os.ReadFile(protoPath)
	if err != nil {
		return "", fmt.Errorf("failed to read proto file: %w", err)
	}

	// Normalize line endings: CRLF/CR → LF so a trailing "\r" doesn't
	// break suffix checks (semicolon, "{", "{}") on rpc lines.
	content := strings.ReplaceAll(string(data), "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	lines := strings.Split(content, "\n")

	// Check if the import already exists
	hasAnnotationImport := false
	for _, line := range lines {
		if strings.TrimSpace(line) == `import "google/api/annotations.proto";` {
			hasAnnotationImport = true
			break
		}
	}

	// Find where to insert the import statement
	importIdx := -1
	if !hasAnnotationImport {
		importIdx = findImportInsertIndex(lines)
	}

	var (
		result      []string
		inService   bool
		serviceName string
	)

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Insert import statement at the determined position
		if i == importIdx {
			result = append(result, `import "google/api/annotations.proto";`)
		}

		// Detect service block start
		if matches := serviceStartRe.FindStringSubmatch(trimmed); len(matches) > 1 && !inService {
			inService = true
			serviceName = matches[1]
			result = append(result, line)
			continue
		}

		// Detect service block end (standalone })
		if inService && trimmed == "}" {
			inService = false
			result = append(result, line)
			continue
		}

		// Transform rpc lines inside service blocks
		if inService {
			if matches := rpcMethodRe.FindStringSubmatch(line); len(matches) > 1 {
				indent := matches[1]
				rpcName := matches[2]

				// Clean the rpc line: remove trailing whitespace/semicolon,
				// and normalize the opening brace. The source rpc line may be:
				//   "rpc Foo(...) returns(...);"   (semicolon style)
				//   "rpc Foo(...) returns(...) {"  (opening brace style)
				//   "rpc Foo(...) returns(...) {}" (empty block style)
				cleanLine := strings.TrimRight(line, " \t")
				cleanLine = strings.TrimSuffix(cleanLine, ";")
				cleanLine = strings.TrimRight(cleanLine, " \t")
				// A trailing "{}" is an empty block — strip the closing brace
				// so it becomes an opening brace for the annotation block.
				if strings.HasSuffix(cleanLine, "{}") {
					cleanLine = cleanLine[:len(cleanLine)-1]
				}
				if !strings.HasSuffix(cleanLine, "{") {
					cleanLine += " {"
				}

				result = append(result, cleanLine)

				// Add the http annotation
				annotation := fmt.Sprintf(`%s  option (google.api.http) = { post: "/%s/%s" body: "*" };`,
					indent, serviceName, rpcName)
				result = append(result, annotation)

				// Add closing brace for the rpc block
				result = append(result, indent+"}")
				continue
			}
		}

		result = append(result, line)
	}

	// Build output path: vehicle.proto → vehicle.kong.proto
	dir := filepath.Dir(protoPath)
	base := filepath.Base(protoPath)
	ext := filepath.Ext(base)
	name := strings.TrimSuffix(base, ext)
	outputPath := filepath.Join(dir, name+".kong"+ext)

	output := strings.Join(result, "\n")
	if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
		return "", fmt.Errorf("failed to write kong proto file: %w", err)
	}

	return outputPath, nil
}

// findImportInsertIndex determines the line index at which to insert
// the import statement for google/api/annotations.proto.
func findImportInsertIndex(lines []string) int {
	lastImportIdx := -1
	goPackageIdx := -1
	syntaxIdx := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "import ") {
			lastImportIdx = i
		}
		if strings.HasPrefix(trimmed, "option go_package") {
			goPackageIdx = i
		}
		if strings.HasPrefix(trimmed, "syntax ") {
			syntaxIdx = i
		}
	}

	if lastImportIdx >= 0 {
		return lastImportIdx + 1
	}
	if goPackageIdx >= 0 {
		return goPackageIdx + 1
	}
	if syntaxIdx >= 0 {
		return syntaxIdx + 1
	}
	return 0
}
