package definition

import (
	"os"
	"path/filepath"
	"reflect"
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

func TestDecodeFile_YAML_stepMetadata(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "with-step-meta.yaml")
	const body = `
schemaVersion: "0.1"
id: meta-decode
version: "1"
initialStepId: a
terminalStepIds: [e]
steps:
  - id: a
    kind: human
    description: "COR capture for SG portrait flow"
    labels:
      - region:sg
      - screen:cor
    annotations:
      product: kyc
      channel: portrait
    presentationRef: sgb/cor@v1
  - id: e
    kind: end
    description: Terminal
    labels: [outcome:done]
transitions:
  - from: a
    to: e
    priority: 0
`
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	def, err := DecodeFile(path)
	if err != nil {
		t.Fatalf("DecodeFile: %v", err)
	}
	if def == nil || def.ID != "meta-decode" {
		t.Fatalf("unexpected def id: %+v", def)
	}
	if len(def.Steps) != 2 {
		t.Fatalf("steps: want 2, got %d", len(def.Steps))
	}
	a := def.Steps[0]
	if a.ID != "a" || a.Kind != StepKindHuman {
		t.Fatalf("step a: %+v", a)
	}
	if want := "COR capture for SG portrait flow"; a.Description != want {
		t.Fatalf("description: want %q, got %q", want, a.Description)
	}
	wantLabels := []string{"region:sg", "screen:cor"}
	if !reflect.DeepEqual(a.Labels, wantLabels) {
		t.Fatalf("labels: want %#v, got %#v", wantLabels, a.Labels)
	}
	wantAnn := map[string]string{
		"product": "kyc",
		"channel": "portrait",
	}
	if !reflect.DeepEqual(a.Annotations, wantAnn) {
		t.Fatalf("annotations: want %#v, got %#v", wantAnn, a.Annotations)
	}
	e := def.Steps[1]
	if e.Description != "Terminal" {
		t.Fatalf("end step description: %q", e.Description)
	}
	if len(e.Labels) != 1 || e.Labels[0] != "outcome:done" {
		t.Fatalf("end labels: %#v", e.Labels)
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
