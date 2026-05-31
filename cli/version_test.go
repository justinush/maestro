package cli

import "testing"

func TestResolvedVersion_ldflags(t *testing.T) {
	t.Parallel()
	old := Version
	t.Cleanup(func() { Version = old })

	Version = "v0.1.1"
	if got := ResolvedVersion(); got != "v0.1.1" {
		t.Fatalf("ResolvedVersion() = %q, want v0.1.1", got)
	}
}

func TestResolvedVersion_devDefault(t *testing.T) {
	t.Parallel()
	old := Version
	t.Cleanup(func() { Version = old })

	Version = "dev"
	got := ResolvedVersion()
	if got != "dev" {
		t.Fatalf("ResolvedVersion() = %q, want dev", got)
	}
}

func TestResolvedVersion_ldflagsOverrideBuildInfo(t *testing.T) {
	t.Parallel()
	old := Version
	t.Cleanup(func() { Version = old })

	Version = "v9.9.9"
	if got := ResolvedVersion(); got != "v9.9.9" {
		t.Fatalf("ldflags should win, got %q", got)
	}
}
