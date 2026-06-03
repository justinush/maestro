package workflow_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/justinush/maestro/pkg/engine"
	"github.com/justinush/maestro/pkg/maestro"
	"github.com/justinush/maestro/pkg/run"
	"github.com/justinush/maestro/pkg/validate"
	"github.com/justinush/maestro/pkg/workflow"
)

const minimalWorkflowYAML = `
schemaVersion: "0.1"
id: wf-a
version: "1"
initialStepId: start
terminalStepIds: [done]
steps:
  - id: start
    kind: action
  - id: done
    kind: end
transitions:
  - from: start
    to: done
    priority: 0
`

const minimalWorkflowYAMLB = `
schemaVersion: "0.1"
id: wf-b
version: "2"
initialStepId: start
terminalStepIds: [done]
steps:
  - id: start
    kind: action
  - id: done
    kind: end
transitions:
  - from: start
    to: done
    priority: 0
`

func loadTestRuntime(t *testing.T, yaml string) *maestro.Runtime {
	t.Helper()
	rt, err := maestro.LoadYAML([]byte(yaml))
	if err != nil {
		t.Fatal(err)
	}
	return rt
}

func writeWorkflowFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	t.Parallel()
	reg := workflow.NewRegistry()
	rt := loadTestRuntime(t, minimalWorkflowYAML)
	if err := reg.Register(rt); err != nil {
		t.Fatal(err)
	}
	err := reg.Register(rt)
	if !errors.Is(err, workflow.ErrDuplicateKey) {
		t.Fatalf("want ErrDuplicateKey, got %v", err)
	}
}

func TestRegistry_RegisterNilRuntime(t *testing.T) {
	t.Parallel()
	reg := workflow.NewRegistry()
	err := reg.Register(nil)
	if err == nil {
		t.Fatal("want error")
	}
	if errors.Is(err, workflow.ErrDuplicateKey) || errors.Is(err, workflow.ErrNotFound) {
		t.Fatalf("unexpected sentinel: %v", err)
	}
}

func TestRegistry_LookupNotFound(t *testing.T) {
	t.Parallel()
	reg := workflow.NewRegistry()
	_, err := reg.Lookup(workflow.Key{ID: "missing", Version: "0"})
	if !errors.Is(err, workflow.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestRegistry_RestoreInstanceNotFound(t *testing.T) {
	t.Parallel()
	reg := workflow.NewRegistry()
	if err := reg.Register(loadTestRuntime(t, minimalWorkflowYAML)); err != nil {
		t.Fatal(err)
	}
	rec := &run.RunRecord{
		RunID:           "run-1",
		WorkflowID:      "wf-a",
		WorkflowVersion: "wrong-version",
		State:           engine.Snapshot{CurrentStepID: "start"},
	}
	_, err := reg.RestoreInstance(rec, maestro.InstanceOptions{})
	if !errors.Is(err, workflow.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	if errors.Is(err, workflow.ErrDefinitionMismatch) {
		t.Fatalf("wrong version should fail at lookup, not mismatch: %v", err)
	}
}

func TestRegistry_NewInstanceAndList(t *testing.T) {
	t.Parallel()
	reg := workflow.NewRegistry()
	if err := reg.Register(loadTestRuntime(t, minimalWorkflowYAML)); err != nil {
		t.Fatal(err)
	}
	if err := reg.Register(loadTestRuntime(t, minimalWorkflowYAMLB)); err != nil {
		t.Fatal(err)
	}

	keys := reg.List()
	if len(keys) != 2 {
		t.Fatalf("List: want 2 keys, got %d", len(keys))
	}
	if keys[0].ID != "wf-a" || keys[1].ID != "wf-b" {
		t.Fatalf("List order: %#v", keys)
	}

	in, err := reg.NewInstance(workflow.Key{ID: "wf-a", Version: "1"}, maestro.InstanceOptions{})
	if err != nil {
		t.Fatal(err)
	}
	res := in.RunUntilBlocked()
	if res.Status != engine.RunCompleted {
		t.Fatalf("RunUntilBlocked: %v err=%v", res.Status, res.Err)
	}
}

func TestRegistry_RestoreInstanceRoundTrip(t *testing.T) {
	t.Parallel()
	reg := workflow.NewRegistry()
	rt := loadTestRuntime(t, minimalWorkflowYAML)
	if err := reg.Register(rt); err != nil {
		t.Fatal(err)
	}

	in, err := reg.NewInstance(workflow.Key{ID: "wf-a", Version: "1"}, maestro.InstanceOptions{RunID: "run-1"})
	if err != nil {
		t.Fatal(err)
	}
	_ = in.RunUntilBlocked()

	rec := run.RecordFromInstance(in, rt.Definition(), 0)
	rec.RunID = "run-1"

	in2, err := reg.RestoreInstance(rec, maestro.InstanceOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if in2.RunID() != "run-1" {
		t.Fatalf("RunID: %q", in2.RunID())
	}
}

func TestLoadDir_Success(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeWorkflowFile(t, dir, "a.yaml", minimalWorkflowYAML)
	writeWorkflowFile(t, dir, "b.yaml", minimalWorkflowYAMLB)

	reg, err := workflow.LoadDir(dir, validate.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !reg.Contains(workflow.Key{ID: "wf-a", Version: "1"}) {
		t.Fatal("missing wf-a")
	}
	if !reg.Contains(workflow.Key{ID: "wf-b", Version: "2"}) {
		t.Fatal("missing wf-b")
	}
}

func TestLoadDir_EmptyDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	reg, err := workflow.LoadDir(dir, validate.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(reg.List()) != 0 {
		t.Fatalf("want empty registry, got %d", len(reg.List()))
	}
}

func TestLoadDir_FailFastInvalidFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeWorkflowFile(t, dir, "01-good.yaml", minimalWorkflowYAML)
	writeWorkflowFile(t, dir, "02-bad.yaml", "not: yaml: [[")

	reg, err := workflow.LoadDir(dir, validate.Options{})
	if err == nil {
		t.Fatal("want error")
	}
	if reg != nil {
		t.Fatal("want nil registry on error")
	}
	if !strings.Contains(err.Error(), "02-bad.yaml") {
		t.Fatalf("want error for second file, got: %v", err)
	}
}

func TestLoadDir_FailFastDuplicate(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeWorkflowFile(t, dir, "01-first.yaml", minimalWorkflowYAML)
	writeWorkflowFile(t, dir, "02-second.yaml", minimalWorkflowYAML)

	_, err := workflow.LoadDir(dir, validate.Options{})
	if !errors.Is(err, workflow.ErrDuplicateKey) {
		t.Fatalf("want ErrDuplicateKey, got %v", err)
	}
	if !strings.Contains(err.Error(), "02-second.yaml") {
		t.Fatalf("want duplicate error on second file, got: %v", err)
	}
}

func TestLoadDir_SkipsUnsupportedExtensions(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	writeWorkflowFile(t, dir, "ok.yaml", minimalWorkflowYAML)
	writeWorkflowFile(t, dir, "readme.txt", "nope")

	reg, err := workflow.LoadDir(dir, validate.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(reg.List()) != 1 {
		t.Fatalf("want 1 workflow, got %d", len(reg.List()))
	}
}
