package main

import (
	"bytes"
	"context"
	_ "embed"
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

	res := in.RunUntilBlocked()
	if res.Status != engine.RunBlocked {
		if res.Err != nil {
			return fmt.Errorf("first run: want RunBlocked, got %v: %w", res.Status, res.Err)
		}
		return fmt.Errorf("first run: want RunBlocked, got %v", res.Status)
	}
	fmt.Printf("blocked at: %s\n", res.StepID)

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
	sub := in.SubmitInput(map[string]any{"fullName": "Demo User"})
	switch sub.Status {
	case engine.SubmitAdvanced:
		// ok
	case engine.SubmitFailed:
		return fmt.Errorf("submit input: %w", sub.Err)
	default:
		return fmt.Errorf("submit input: expected SubmitAdvanced from collect-profile, got %v", sub.Status)
	}
	fmt.Printf("continued to: %s\n", sub.StepID)

	for {
		res := in.RunUntilBlocked()
		switch res.Status {
		case engine.RunBlocked:
			return fmt.Errorf("unexpected block at step %q", res.StepID)
		case engine.RunCompleted:
			fmt.Printf("completed at: %s\n", res.StepID)
			if err := persistFinal(ctx, store, in, def); err != nil {
				return err
			}
			fmt.Println()
			fmt.Println("trace:")
			for _, ev := range in.Events() {
				fmt.Println(ev.String())
			}
			return nil
		case engine.RunFailed:
			return fmt.Errorf("run: %w", res.Err)
		default:
			return fmt.Errorf("run: unexpected status %v", res.Status)
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
