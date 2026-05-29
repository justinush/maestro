package postgres_test

import (
	"context"
	"errors"
	"os"
	"sync"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/justinush/maestro/pkg/definition"
	"github.com/justinush/maestro/pkg/engine"
	"github.com/justinush/maestro/pkg/run"
	"github.com/justinush/maestro/pkg/run/postgres"
)

var (
	testPool     *pgxpool.Pool
	testPoolOnce sync.Once
	testPoolErr  error
)

func testStore(t *testing.T) *postgres.Store {
	t.Helper()

	dsn := os.Getenv("MAESTRO_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set MAESTRO_TEST_DATABASE_URL to run Postgres store integration tests")
	}

	testPoolOnce.Do(func() {
		ctx := context.Background()
		testPool, testPoolErr = pgxpool.New(ctx, dsn)
		if testPoolErr != nil {
			return
		}
		testPoolErr = postgres.ApplySchema(ctx, testPool)
	})
	if testPoolErr != nil {
		t.Fatalf("test pool: %v", testPoolErr)
	}

	ctx := context.Background()
	if _, err := testPool.Exec(ctx, `TRUNCATE workflow_runs`); err != nil {
		t.Fatalf("truncate workflow_runs: %v", err)
	}
	return postgres.NewStore(testPool)
}

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

func TestStore_CreateGet(t *testing.T) {
	ctx := context.Background()
	st := testStore(t)

	def := minimalCompletedWorkflow()
	in, err := engine.NewInstance(def, engine.Options{RunID: "run-1"})
	if err != nil {
		t.Fatal(err)
	}
	res := in.RunUntilBlocked()
	if res.Status != engine.RunCompleted {
		t.Fatalf("RunUntilBlocked: want RunCompleted, got %v err=%v", res.Status, res.Err)
	}

	rec := run.RecordFromInstance(in, def, 0)
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

func TestStore_CreateDuplicate(t *testing.T) {
	ctx := context.Background()
	st := testStore(t)

	rec := &run.RunRecord{RunID: "x", WorkflowID: "w", WorkflowVersion: "1", State: engine.Snapshot{CurrentStepID: "a"}}
	if err := st.Create(ctx, rec); err != nil {
		t.Fatal(err)
	}
	if err := st.Create(ctx, rec); !errors.Is(err, run.ErrExists) {
		t.Fatalf("want ErrExists, got %v", err)
	}
}

func TestStore_GetNotFound(t *testing.T) {
	ctx := context.Background()
	st := testStore(t)

	_, err := st.Get(ctx, "missing")
	if !errors.Is(err, run.ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	_, err = st.Get(ctx, "")
	if !errors.Is(err, run.ErrNotFound) {
		t.Fatalf("empty id: want ErrNotFound, got %v", err)
	}
}

func TestStore_SaveIncrementRevision(t *testing.T) {
	ctx := context.Background()
	st := testStore(t)

	rec := &run.RunRecord{RunID: "r", WorkflowID: "w", WorkflowVersion: "1", State: engine.Snapshot{CurrentStepID: "a"}}
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

func TestStore_SaveRevisionConflict(t *testing.T) {
	ctx := context.Background()
	st := testStore(t)

	rec := &run.RunRecord{RunID: "r", WorkflowID: "w", WorkflowVersion: "1", State: engine.Snapshot{CurrentStepID: "a"}}
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
	if err := st.Save(ctx, second); !errors.Is(err, run.ErrRevisionConflict) {
		t.Fatalf("want ErrRevisionConflict, got %v", err)
	}
}

func TestStore_InstanceFromRecordRoundTrip(t *testing.T) {
	ctx := context.Background()
	st := testStore(t)

	def := minimalCompletedWorkflow()
	in, err := engine.NewInstance(def, engine.Options{RunID: "run-1", InitialVariables: map[string]any{"n": 42}})
	if err != nil {
		t.Fatal(err)
	}
	res := in.RunUntilBlocked()
	if res.Status != engine.RunCompleted {
		t.Fatalf("RunUntilBlocked: want RunCompleted, got %v err=%v", res.Status, res.Err)
	}

	rec := run.RecordFromInstance(in, def, 0)
	rec.RunID = "run-1"
	if err := st.Create(ctx, rec); err != nil {
		t.Fatal(err)
	}
	loaded, err := st.Get(ctx, "run-1")
	if err != nil {
		t.Fatal(err)
	}

	in2, err := run.InstanceFromRecord(loaded, def, engine.Options{ActionRegistry: engine.DefaultRegistry()})
	if err != nil {
		t.Fatalf("InstanceFromRecord: %v", err)
	}
	if in2.CurrentStepID() != "end" {
		t.Fatalf("step: %q", in2.CurrentStepID())
	}
	if in2.RunID() != "run-1" {
		t.Fatalf("runId: %q", in2.RunID())
	}
	v := in2.Variables()["n"]
	switch v.(type) {
	case int, int64, float64:
	default:
		t.Fatalf("variables n type %T value %#v", v, v)
	}
}
