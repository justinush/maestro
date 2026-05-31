package cli

import "runtime/debug"

var Version = "dev"

func ResolvedVersion() string {
	if Version != "" && Version != "dev" {
		return Version
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return Version
	}
	if v := info.Main.Version; v != "" && v != "(devel)" {
		return v
	}
	return Version
}
