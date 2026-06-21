package validate

import (
	"strings"
	"testing"
)

func TestNormalizeAllowedActionTypes(t *testing.T) {
	t.Parallel()

	_, err := normalizeAllowedActionTypes([]string{
		"vendor-create-session",
		" sumsub-create-check ",
		"vendor-create-session",
	})
	if err == nil {
		t.Fatal("expected duplicate error")
	}

	got, err := normalizeAllowedActionTypes([]string{
		"vendor-create-session",
		"sumsub-create-check",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "sumsub-create-check" || got[1] != "vendor-create-session" {
		t.Fatalf("got %v", got)
	}

	_, err = normalizeAllowedActionTypes([]string{"stub"})
	if err == nil || !strings.Contains(err.Error(), "built-in") {
		t.Fatalf("want built-in error, got %v", err)
	}

	_, err = normalizeAllowedActionTypes([]string{"Vendor_Create"})
	if err == nil {
		t.Fatal("expected pattern error")
	}
}

func TestExtendActionTypeEnum(t *testing.T) {
	t.Parallel()

	schemaDoc := map[string]any{
		"$defs": map[string]any{
			"action": map[string]any{
				"properties": map[string]any{
					"type": map[string]any{
						"enum": []any{"stub", "http"},
					},
				},
			},
		},
	}

	if err := extendActionTypeEnum(schemaDoc, []string{"vendor-create-session"}); err != nil {
		t.Fatal(err)
	}

	defs, ok := schemaDoc["$defs"].(map[string]any)
	if !ok {
		t.Fatal("$defs is not map[string]any")
	}
	action, ok := defs["action"].(map[string]any)
	if !ok {
		t.Fatal("action is not map[string]any")
	}
	props, ok := action["properties"].(map[string]any)
	if !ok {
		t.Fatal("properties is not map[string]any")
	}
	typeProp, ok := props["type"].(map[string]any)
	if !ok {
		t.Fatal("type is not map[string]any")
	}
	enum, ok := typeProp["enum"].([]any)
	if !ok {
		t.Fatal("enum is not []any")
	}
	if len(enum) != 3 {
		t.Fatalf("enum len: got %d", len(enum))
	}
}
