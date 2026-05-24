// Package maestro is the canonical embedding API for Maestro.
//
// Typical application flow:
//
//	rt, err := maestro.Load("workflow.yaml")
//	in, err := rt.NewInstance(maestro.InstanceOptions{...})
//	res := in.RunUntilBlocked()
//	// when res.Status == engine.RunBlocked:
//	sub := in.SubmitInput(map[string]any{...})
//
// Persist and resume across requests with pkg/run.Store, run.RecordFromInstance,
// and rt.RestoreInstance (preferred over run.InstanceFromRecord in application code).
//
// Use pkg/engine and pkg/run directly for advanced or custom integrations.
package maestro

import (
	"fmt"

	"github.com/justinush/maestro/pkg/definition"
	"github.com/justinush/maestro/pkg/engine"
	"github.com/justinush/maestro/pkg/run"
	"github.com/justinush/maestro/pkg/validate"
)

// Runtime is a validated workflow definition ready to instantiate with Runtime.NewInstance.
type Runtime struct {
	def *definition.WorkflowDefinition
}

// Load reads path, decodes it, and validates with default validate.Options.
func Load(path string) (*Runtime, error) {
	return loadFromPath(path, validate.Options{})
}

// LoadWithValidate is like Load but uses custom validation options (schema path, verbose, etc.).
func LoadWithValidate(path string, vopts validate.Options) (*Runtime, error) {
	return loadFromPath(path, vopts)
}

func loadFromPath(path string, vopts validate.Options) (*Runtime, error) {
	def, err := definition.DecodeFile(path)
	if err != nil {
		return nil, fmt.Errorf("maestro load: decode: %w", err)
	}
	return compileDefinition(def, vopts)
}

// LoadYAML decodes and validates workflow YAML bytes (for //go:embed, config services, tests).
func LoadYAML(data []byte) (*Runtime, error) {
	return LoadYAMLWithValidate(data, validate.Options{})
}

// LoadYAMLWithValidate is LoadYAML with custom validate.Options.
func LoadYAMLWithValidate(data []byte, vopts validate.Options) (*Runtime, error) {
	def, err := definition.DecodeYAML(data)
	if err != nil {
		return nil, fmt.Errorf("maestro load yaml: decode: %w", err)
	}
	return compileDefinition(def, vopts)
}

// LoadJSON decodes and validates workflow JSON bytes.
func LoadJSON(data []byte) (*Runtime, error) {
	return LoadJSONWithValidate(data, validate.Options{})
}

// LoadJSONWithValidate is LoadJSON with custom validate.Options.
func LoadJSONWithValidate(data []byte, vopts validate.Options) (*Runtime, error) {
	def, err := definition.DecodeJSON(data)
	if err != nil {
		return nil, fmt.Errorf("maestro load json: decode: %w", err)
	}
	return compileDefinition(def, vopts)
}

// Compile validates an in-memory definition (codegen, tests) without reading bytes from disk.
func Compile(def *definition.WorkflowDefinition, vopts validate.Options) (*Runtime, error) {
	if def == nil {
		return nil, fmt.Errorf("maestro compile: nil definition")
	}
	return compileDefinition(def, vopts)
}

func compileDefinition(def *definition.WorkflowDefinition, vopts validate.Options) (*Runtime, error) {
	if err := validate.WorkflowDefinition(def, vopts); err != nil {
		return nil, fmt.Errorf("maestro: validate: %w", err)
	}
	return &Runtime{def: def}, nil
}

// Definition returns the validated workflow. Callers must not mutate it.
func (rt *Runtime) Definition() *definition.WorkflowDefinition {
	if rt == nil {
		return nil
	}
	return rt.def
}

// InstanceOptions configures Runtime.NewInstance and Runtime.RestoreInstance.
// A zero value uses DefaultRegistry (stub only).
type InstanceOptions struct {
	// RunID overrides the run id; when empty on restore, RunRecord.RunID is used.
	RunID string

	// InitialVariables are applied only when creating a new instance (not on restore).
	InitialVariables map[string]any

	// TraceGuards enables transition guard events in the execution trace.
	TraceGuards bool

	// ActionRegistry provides action runners; nil uses engine.DefaultRegistry().
	ActionRegistry *engine.Registry
}

// NewInstance builds an engine.Instance for this workflow at initialStepId.
func (rt *Runtime) NewInstance(opts InstanceOptions) (*engine.Instance, error) {
	if rt == nil || rt.def == nil {
		return nil, engine.ErrNilDefinition
	}
	return engine.NewInstance(rt.def, instanceOptionsToEngine(opts))
}

// RestoreInstance rebuilds an engine.Instance from a stored RunRecord.
//
// This is the preferred restore path for application code. Use the same
// ActionRegistry and TraceGuards semantics as when the run was created.
// RunID is taken from opts when set, otherwise from rec.RunID.
//
// run.InstanceFromRecord offers the same restore at the persistence layer for
// custom Store implementations.
func (rt *Runtime) RestoreInstance(rec *run.RunRecord, opts InstanceOptions) (*engine.Instance, error) {
	if rt == nil || rt.def == nil {
		return nil, engine.ErrNilDefinition
	}
	return run.InstanceFromRecord(rec, rt.def, instanceOptionsToEngine(opts))
}

func instanceOptionsToEngine(opts InstanceOptions) engine.Options {
	reg := opts.ActionRegistry
	if reg == nil {
		reg = engine.DefaultRegistry()
	}
	return engine.Options{
		RunID:            opts.RunID,
		InitialVariables: opts.InitialVariables,
		TraceGuards:      opts.TraceGuards,
		ActionRegistry:   reg,
	}
}
