package definition

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDecodeFile_JSON_valid(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "workflow.json")
	const body = `{
  "schemaVersion": "0.1",
  "id": "x",
  "version": "1",
  "initialStepId": "a",
  "terminalStepIds": ["e"],
  "steps": [
    {"id": "a", "kind": "human", "presentationRef": "p"},
    {"id": "e", "kind": "end"}
  ],
  "transitions": [
    {"from": "a", "to": "e", "priority": 0}
  ]
}`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	def, err := DecodeFile(path)
	if err != nil {
		t.Fatalf("DecodeFile: %v", err)
	}
	if def == nil || def.ID != "x" {
		t.Fatalf("unexpected def: %+v", def)
	}
}

func TestDecodeFile_JSON_trailingContent(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	body := `{
  "schemaVersion": "0.1",
  "id": "x",
  "version": "1",
  "initialStepId": "a",
  "terminalStepIds": ["e"],
  "steps": [
    {"id": "a", "kind": "human", "presentationRef": "p"},
	{"id": "e", "kind": "end" }
  ],
  "transitions": [
    {"from":"a","to":"e","priority":0}
  ]
}
{"extra":true}
`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := DecodeFile(path)
	if err == nil {
		t.Fatal("expected error for trailing JSON")
	}
	if !strings.Contains(err.Error(), "trailing content") {
		t.Fatalf("error %q should mention trailing content", err.Error())
	}
}

func TestDecodeFile_JSON_unknownField(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "unknown.json")
	const body = `{
  "schemaVersion": "0.1",
  "id": "x",
  "version": "1",
  "initialStepId": "a",
  "terminalStepIds": ["e"],
  "unknownRoot": true,
  "steps": [
    {"id": "a", "kind": "human", "presentationRef": "p"},
    {"id": "e", "kind": "end"}
  ],
  "transitions": [
    {"from": "a", "to": "e", "priority": 0}
  ]
}
`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := DecodeFile(path)
	if err == nil {
		t.Fatal("expected error for unknown JSON field")
	}
}

func TestDecodeFile_unsupportedExtension(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "workflow.txt")
	if err := os.WriteFile(path, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := DecodeFile(path)
	if err == nil || !strings.Contains(err.Error(), "unsupported file extension") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDecodeFile_YAML_unknownField(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "unknown.yaml")
	const body = `
schemaVersion: "0.1"
id: x
version: "1"
unknownRoot: true
initialStepId: a
terminalStepIds: [e]
steps:
  - id: a
    kind: human
    presentationRef: p
  - id: e
    kind: end
transitions:
  - from: a
    to: e
    priority: 0
`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := DecodeFile(path)
	if err == nil {
		t.Fatal("expected error for unknown YAML field")
	}
}
