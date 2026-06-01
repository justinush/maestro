package cli

import "runtime/debug"

var Version = "dev"

func ResolvedVersion() string {
	return resolveVersion(Version)
}

func resolveVersion(ver string) string {
	if ver != "" && ver != "dev" {
		return ver
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return ver
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		return v
	}
	return ver
}
