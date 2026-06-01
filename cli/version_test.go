package cli

import "testing"

func TestResolveVersion_ldflags(t *testing.T) {
	t.Parallel()
	if got := resolveVersion("v0.1.1"); got != "v0.1.1" {
		t.Fatalf("resolveVersion() = %q, want v0.1.1", got)
	}
}

func TestResolveVersion_devDefault(t *testing.T) {
	t.Parallel()
	got := resolveVersion("dev")
	if got != "dev" {
		t.Fatalf("resolveVersion() = %q, want dev", got)
	}
}

func TestResolveVersion_ldflagsOverrideBuildInfo(t *testing.T) {
	t.Parallel()
	if got := resolveVersion("v9.9.9"); got != "v9.9.9" {
		t.Fatalf("ldflags should win, got %q", got)
	}
}
