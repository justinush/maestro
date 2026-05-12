package main

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"

	"github.com/justinush/maestro/pkg/definition"
	"github.com/justinush/maestro/pkg/engine"
	"github.com/justinush/maestro/pkg/run"
	"github.com/justinush/maestro/pkg/validate"
	"gopkg.in/yaml.v3"
)

//go:embed workflow.yaml
var embeddedWorkflow []byte

const demoRunID = "run_demo"

func runDemo() error {
	ctx := context.Background()

	def, err := decodeEmbeddedWorkflow()
	if err != nil {
		return err
	}
	if err := validate.WorkflowDefinition(def, validate.Options{}); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	store := run.NewMemoryStore()

	fmt.Printf("created run: %s\n", demoRunID)

	in, err := engine.NewInstance(def, engine.Options{
		RunID:            demoRunID,
		ActionRegistry:   engine.DefaultRegistry(),
		InitialVariables: map[string]any{},
	})
	if err != nil {
		return fmt.Errorf("new instance: %w", err)
	}

	if err := in.RunUntilBlocked(); err != nil {
		if !isErrNeedsInput(err) {
			return fmt.Errorf("first run: %w", err)
		}
	}
	fmt.Printf("blocked at: %s\n", in.CurrentStepID())

	rec := run.RecordFromInstance(in, def, 0)
	if rec == nil {
		return fmt.Errorf("record from instance: nil")
	}
	if err := store.Create(ctx, rec); err != nil {
		return fmt.Errorf("store create: %w", err)
	}

	// Next handler only uses persisted state: reload from the store (same as a new HTTP request).
	loaded, err := store.Get(ctx, demoRunID)
	if err != nil {
		return fmt.Errorf("store get: %w", err)
	}

	in, err = run.InstanceFromRecord(loaded, def, engine.Options{
		ActionRegistry: engine.DefaultRegistry(),
	})
	if err != nil {
		return fmt.Errorf("restore instance: %w", err)
	}

	fmt.Println()
	fmt.Println("submitting profile input...")
	advanced, err := in.SubmitInput(map[string]any{"fullName": "Demo User"})
	if err != nil {
		return fmt.Errorf("submit input: %w", err)
	}
	if !advanced {
		return fmt.Errorf("submit input: expected to advance from collect-profile")
	}
	fmt.Printf("continued to: %s\n", in.CurrentStepID())

	for {
		err := in.RunUntilBlocked()
		switch {
		case isErrNeedsInput(err):
			return fmt.Errorf("unexpected block at step %q", in.CurrentStepID())
		case isErrWorkflowCompleted(err):
			fmt.Printf("completed at: %s\n", in.CurrentStepID())
			if err := persistFinal(ctx, store, in, def); err != nil {
				return err
			}
			fmt.Println()
			fmt.Println("trace:")
			for _, ev := range in.Events() {
				fmt.Println(ev.String())
			}
			return nil
		default:
			return fmt.Errorf("run: %w", err)
		}
	}
}

func decodeEmbeddedWorkflow() (*definition.WorkflowDefinition, error) {
	dec := yaml.NewDecoder(bytes.NewReader(embeddedWorkflow))
	dec.KnownFields(true)
	var def definition.WorkflowDefinition
	if err := dec.Decode(&def); err != nil {
		return nil, fmt.Errorf("parse embedded workflow: %w", err)
	}
	return &def, nil
}

func persistFinal(ctx context.Context, store run.Store, in *engine.Instance, def *definition.WorkflowDefinition) error {
	loaded, err := store.Get(ctx, demoRunID)
	if err != nil {
		return fmt.Errorf("store get before final save: %w", err)
	}
	rec := run.RecordFromInstance(in, def, loaded.Revision)
	if rec == nil {
		return fmt.Errorf("record from instance: nil")
	}
	if err := store.Save(ctx, rec); err != nil {
		return fmt.Errorf("final save: %w", err)
	}
	return nil
}

func isErrNeedsInput(err error) bool {
	return errors.Is(err, engine.ErrNeedsInput)
}

func isErrWorkflowCompleted(err error) bool {
	return errors.Is(err, engine.ErrWorkflowCompleted)
}
