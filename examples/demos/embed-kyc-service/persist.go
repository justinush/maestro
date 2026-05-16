package main

// These helpers mimic what a real service would do with a database-backed run.Store.

import (
	"context"
	"fmt"

	"github.com/justinush/maestro/pkg/definition"
	"github.com/justinush/maestro/pkg/engine"
	"github.com/justinush/maestro/pkg/maestro"
	"github.com/justinush/maestro/pkg/run"
)

func persistNewRun(ctx context.Context, store run.Store, in *engine.Instance, def *definition.WorkflowDefinition) error {
	rec := run.RecordFromInstance(in, def, 0)
	if rec == nil {
		return fmt.Errorf("record from instance: nil")
	}
	if err := store.Create(ctx, rec); err != nil {
		return fmt.Errorf("store create: %w", err)
	}
	return nil
}

func saveRun(ctx context.Context, store run.Store, runID string, in *engine.Instance, def *definition.WorkflowDefinition) error {
	loaded, err := store.Get(ctx, runID)
	if err != nil {
		return fmt.Errorf("store get before save: %w", err)
	}
	rec := run.RecordFromInstance(in, def, loaded.Revision)
	if rec == nil {
		return fmt.Errorf("record from instance: nil")
	}
	if err := store.Save(ctx, rec); err != nil {
		return fmt.Errorf("store save: %w", err)
	}
	return nil
}

func restoreRun(ctx context.Context, rt *maestro.Runtime, store run.Store, runID string) (*engine.Instance, error) {
	loaded, err := store.Get(ctx, runID)
	if err != nil {
		return nil, fmt.Errorf("store get: %w", err)
	}
	in, err := rt.RestoreInstance(loaded, maestro.InstanceOptions{})
	if err != nil {
		return nil, fmt.Errorf("restore instance: %w", err)
	}
	return in, nil
}
