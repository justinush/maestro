package run_test

import (
	"context"
	"fmt"
	"log"

	"github.com/justinush/maestro/pkg/definition"
	"github.com/justinush/maestro/pkg/engine"
	"github.com/justinush/maestro/pkg/run"
)

func ExampleRecordFromInstance() {
	def := &definition.WorkflowDefinition{
		SchemaVersion:   "0.1",
		ID:              "w",
		Version:         "1",
		InitialStepID:   "a",
		TerminalStepIDs: []string{"end"},
		Steps: []definition.Step{
			{ID: "a", Kind: definition.StepKindAction},
			{ID: "end", Kind: definition.StepKindEnd},
		},
		Transitions: []definition.Transition{{From: "a", To: "end"}},
	}
	in, err := engine.NewInstance(def, engine.Options{RunID: "run-1"})
	if err != nil {
		log.Fatal(err)
	}
	if res := in.RunUntilBlocked(); res.Status != engine.RunCompleted {
		log.Fatalf("status: %v err=%v", res.Status, res.Err)
	}

	st := run.NewMemoryStore()
	rec := run.RecordFromInstance(in, def, 0)
	rec.RunID = "run-1"
	if err := st.Create(context.Background(), rec); err != nil {
		log.Fatal(err)
	}
	got, err := st.Get(context.Background(), "run-1")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(got.Revision, got.State.CurrentStepID)
	// Output: 1 end
}
