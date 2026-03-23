package env

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/lerity-yao/czt-contrib/cztctl/internal/version"
	"github.com/lerity-yao/czt-contrib/cztctl/util/pathx"
	"github.com/lerity-yao/czt-contrib/cztctl/vars"
)

const (
	CztctlOS           = "CZTCTL_OS"
	CztctlArch         = "CZTCTL_ARCH"
	CztctlHome         = "CZTCTL_HOME"
	CztctlCache        = "CZTCTL_CACHE"
	CztctlVersion      = "CZTCTL_VERSION"
	CztctlExperimental = "CZTCTL_EXPERIMENTAL"

	envFileDir      = "env"
	ExperimentalOn  = "on"
	ExperimentalOff = "off"
)

// orderedEnv is a lightweight ordered key-value store for env management.
type orderedEnv struct {
	keys   []string
	values map[string]string
}

func newOrderedEnv() *orderedEnv {
	return &orderedEnv{values: make(map[string]string)}
}

func (e *orderedEnv) set(key, value string) {
	if _, exists := e.values[key]; !exists {
		e.keys = append(e.keys, key)
	}
	e.values[key] = value
}

func (e *orderedEnv) get(key string) (string, bool) {
	v, ok := e.values[key]
	return v, ok
}

func (e *orderedEnv) getOr(key, def string) string {
	if v, ok := e.values[key]; ok {
		return v
	}
	return def
}

func (e *orderedEnv) hasKey(key string) bool {
	_, ok := e.values[key]
	return ok
}

func (e *orderedEnv) format() []string {
	var lines []string
	for _, k := range e.keys {
		lines = append(lines, fmt.Sprintf("%s=%s", k, e.values[k]))
	}
	return lines
}

var cztctlEnv *orderedEnv

func init() {
	defaultHome, err := pathx.GetDefaultGoctlHome()
	if err != nil {
		log.Fatalln(err)
	}

	cztctlEnv = newOrderedEnv()
	cztctlEnv.set(CztctlOS, runtime.GOOS)
	cztctlEnv.set(CztctlArch, runtime.GOARCH)

	existsEnv := readEnv(defaultHome)
	if existsEnv != nil {
		if home, ok := existsEnv.get(CztctlHome); ok && len(home) > 0 {
			cztctlEnv.set(CztctlHome, home)
		}
		if cache, ok := existsEnv.get(CztctlCache); ok && len(cache) > 0 {
			cztctlEnv.set(CztctlCache, cache)
		}
		cztctlEnv.set(CztctlExperimental, existsEnv.getOr(CztctlExperimental, ExperimentalOff))
	}

	if !cztctlEnv.hasKey(CztctlHome) {
		cztctlEnv.set(CztctlHome, defaultHome)
	}
	if !cztctlEnv.hasKey(CztctlCache) {
		cacheDir, _ := pathx.GetCacheDir()
		cztctlEnv.set(CztctlCache, cacheDir)
	}
	if !cztctlEnv.hasKey(CztctlExperimental) {
		cztctlEnv.set(CztctlExperimental, ExperimentalOff)
	}

	cztctlEnv.set(CztctlVersion, version.BuildVersion)
}

// Print returns formatted env output. If args are given, only those keys are printed.
func Print(args ...string) string {
	if len(args) == 0 {
		return strings.Join(cztctlEnv.format(), "\n")
	}
	var values []string
	for _, key := range args {
		value, ok := cztctlEnv.get(key)
		if !ok {
			value = fmt.Sprintf("%%not found%%")
		}
		values = append(values, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(values, "\n")
}

// Get returns the value for the given key.
func Get(key string) string {
	return GetOr(key, "")
}

// GetOr returns the value for the given key, or the default if not found.
func GetOr(key, def string) string {
	return cztctlEnv.getOr(key, def)
}

// UseExperimental returns true if CZTCTL_EXPERIMENTAL is "on".
// Default is "off" (ANTLR4 parser). Set to "on" to use the hand-written parser.
func UseExperimental() bool {
	return GetOr(CztctlExperimental, ExperimentalOff) == ExperimentalOn
}

// WriteEnv writes the given key=value pairs to the env file.
func WriteEnv(kv []string) error {
	defaultHome, err := pathx.GetDefaultGoctlHome()
	if err != nil {
		log.Fatalln(err)
	}

	for _, expr := range kv {
		parts := strings.SplitN(expr, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("[writeEnv]: invalid expression: %q, expected KEY=VALUE", expr)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		switch key {
		case CztctlHome, CztctlCache:
			if !pathx.FileExists(value) {
				return fmt.Errorf("[writeEnv]: path %q does not exist", value)
			}
		}

		if !cztctlEnv.hasKey(key) {
			return fmt.Errorf("[writeEnv]: invalid key: %s", key)
		}
		cztctlEnv.set(key, value)
	}

	envFile := filepath.Join(defaultHome, envFileDir)
	return os.WriteFile(envFile, []byte(strings.Join(cztctlEnv.format(), "\n")), 0o777)
}

func readEnv(home string) *orderedEnv {
	envFile := filepath.Join(home, envFileDir)
	data, err := os.ReadFile(envFile)
	if err != nil {
		return nil
	}
	env := newOrderedEnv()
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			env.set(strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]))
		}
	}
	return env
}

// LookPath searches for an executable named file in the
// directories named by the PATH environment variable.
func LookPath(xBin string) (string, error) {
	suffix := getExeSuffix()
	if len(suffix) > 0 && !strings.HasSuffix(xBin, suffix) {
		xBin = xBin + suffix
	}

	bin, err := exec.LookPath(xBin)
	if err != nil {
		return "", err
	}
	return bin, nil
}

// CanExec reports whether the current system can start new processes
// using os.StartProcess or (more commonly) exec.Command.
func CanExec() bool {
	switch runtime.GOOS {
	case vars.OsJs, vars.OsIOS:
		return false
	default:
		return true
	}
}

func getExeSuffix() string {
	if runtime.GOOS == vars.OsWindows {
		return ".exe"
	}
	return ""
}
