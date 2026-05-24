package engine

import iengine "github.com/justinush/maestro/internal/engine"

// Options configures a new workflow instance (identity, initial variables, tracing, action runners).
type Options struct {
	// RunID is an optional correlation id stored on the instance and in snapshots.
	RunID string

	// InitialVariables are shallow-copied into the instance at creation (nested values may still alias).
	InitialVariables map[string]any

	// TraceGuards records transition guard evaluation in the event trace.
	TraceGuards bool

	// ActionRegistry provides runners for workflow action types. When nil, DefaultRegistry is used.
	ActionRegistry *Registry
}

func (o Options) toInternal() iengine.Options {
	var reg *iengine.Registry
	if o.ActionRegistry != nil {
		reg = o.ActionRegistry.impl
	}
	return iengine.Options{
		RunID:            o.RunID,
		InitialVariables: o.InitialVariables,
		TraceGuards:      o.TraceGuards,
		ActionRegistry:   reg,
	}
}
