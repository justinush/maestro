package run

import (
	"context"
	"errors"
	"testing"

	"github.com/justinush/maestro/internal/definition"
	"github.com/justinush/maestro/pkg/engine"
)

func minimalCompletedWorkflow() *definition.WorkflowDefinition {
	return &definition.WorkflowDefinition{
		SchemaVersion:   "0.1",
		ID:              "w",
		Version:         "1",
		InitialStepID:   "a",
		TerminalStepIDs: []string{"end"},
		Steps: []definition.Step{
			{ID: "a", Kind: definition.StepKindAction},
			{ID: "end", Kind: definition.StepKindEnd},
		},
		Transitions: []definition.Transition{
			{From: "a", To: "end", Priority: 0},
		},
	}
}

func TestMemoryStore_CreateGet(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := NewMemoryStore()

	def := minimalCompletedWorkflow()
	in, err := engine.NewInstance(def, engine.Options{RunID: "run-1"})
	if err != nil {
		t.Fatal(err)
	}
	if err := in.RunUntilBlocked(); !errors.Is(err, engine.ErrWorkflowCompleted) {
		t.Fatalf("RunUntilBlocked: %v", err)
	}

	rec := RecordFromInstance(in, def, 0)
	rec.RunID = "run-1"
	if err := st.Create(ctx, rec); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := st.Get(ctx, "run-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Revision != 1 {
		t.Fatalf("revision: want 1, got %d", got.Revision)
	}
	if got.WorkflowID != "w" || got.WorkflowVersion != "1" {
		t.Fatalf("metadata: %#v", got)
	}
	if got.State.CurrentStepID != "end" {
		t.Fatalf("state step: %q", got.State.CurrentStepID)
	}
}

func TestMemoryStore_CreateDuplicate(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := NewMemoryStore()

	rec := &RunRecord{RunID: "x", WorkflowID: "w", WorkflowVersion: "1"}
	if err := st.Create(ctx, rec); err != nil {
		t.Fatal(err)
	}
	if err := st.Create(ctx, rec); !errors.Is(err, ErrExists) {
		t.Fatalf("want ErrExists, got %v", err)
	}
}

func TestMemoryStore_GetNotFound(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := NewMemoryStore()

	_, err := st.Get(ctx, "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	_, err = st.Get(ctx, "")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("empty id: want ErrNotFound, got %v", err)
	}
}

func TestMemoryStore_SaveIncrementsRevision(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := NewMemoryStore()

	rec := &RunRecord{RunID: "r", WorkflowID: "w", WorkflowVersion: "1", State: engine.Snapshot{CurrentStepID: "a"}}
	if err := st.Create(ctx, rec); err != nil {
		t.Fatal(err)
	}
	got, err := st.Get(ctx, "r")
	if err != nil {
		t.Fatal(err)
	}
	got.State.Variables = map[string]any{"k": "v"}
	if err := st.Save(ctx, got); err != nil {
		t.Fatalf("Save: %v", err)
	}
	after, err := st.Get(ctx, "r")
	if err != nil {
		t.Fatal(err)
	}
	if after.Revision != 2 {
		t.Fatalf("revision: want 2, got %d", after.Revision)
	}
	if after.State.Variables["k"] != "v" {
		t.Fatalf("variables: %#v", after.State.Variables)
	}
}

func TestMemoryStore_SaveRevisionConflict(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := NewMemoryStore()

	rec := &RunRecord{RunID: "r", WorkflowID: "w", WorkflowVersion: "1", State: engine.Snapshot{CurrentStepID: "a"}}
	if err := st.Create(ctx, rec); err != nil {
		t.Fatal(err)
	}
	first, err := st.Get(ctx, "r")
	if err != nil {
		t.Fatal(err)
	}
	second, err := st.Get(ctx, "r")
	if err != nil {
		t.Fatal(err)
	}
	first.State.Variables = map[string]any{"from": "first"}
	if err := st.Save(ctx, first); err != nil {
		t.Fatal(err)
	}
	second.State.Variables = map[string]any{"from": "second"}
	if err := st.Save(ctx, second); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("want ErrRevisionConflict, got %v", err)
	}
}

func TestMemoryStore_InstanceFromRecordRoundTrip(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	st := NewMemoryStore()

	def := minimalCompletedWorkflow()
	in, err := engine.NewInstance(def, engine.Options{RunID: "run-1", InitialVariables: map[string]any{"n": 42}})
	if err != nil {
		t.Fatal(err)
	}
	if err := in.RunUntilBlocked(); !errors.Is(err, engine.ErrWorkflowCompleted) {
		t.Fatalf("RunUntilBlocked: %v", err)
	}

	rec := RecordFromInstance(in, def, 0)
	rec.RunID = "run-1"
	if err := st.Create(ctx, rec); err != nil {
		t.Fatal(err)
	}
	loaded, err := st.Get(ctx, "run-1")
	if err != nil {
		t.Fatal(err)
	}

	in2, err := InstanceFromRecord(loaded, def, engine.Options{ActionRegistry: engine.DefaultRegistry()})
	if err != nil {
		t.Fatalf("InstanceFromRecord: %v", err)
	}
	if in2.CurrentStepID() != "end" {
		t.Fatalf("step: %q", in2.CurrentStepID())
	}
	if in2.RunID() != "run-1" {
		t.Fatalf("runId: %q", in2.RunID())
	}
	// JSON round-trip via cloneRecord may turn numbers into float64
	v := in2.Variables()["n"]
	switch v.(type) {
	case int, int64, float64:
		// ok
	default:
		t.Fatalf("variables n type %T value %#v", v, v)
	}
}
