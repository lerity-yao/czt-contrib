package version

import "runtime/debug"

const (
	BuildVersion = "1.10.5"
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
