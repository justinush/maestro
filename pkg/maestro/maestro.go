package maestro

import (
	"fmt"

	"github.com/justinush/maestro/pkg/definition"
	"github.com/justinush/maestro/pkg/engine"
	"github.com/justinush/maestro/pkg/validate"
)

// Runtime is a decoded, validated workflow ready to create engine instances.
// It is a small facade over pkg/definition + pkg/validate for first-touch DX.
type Runtime struct {
	def *definition.WorkflowDefinition
}

// Load reads a workflow file from disk and validates it with default options.
func Load(path string) (*Runtime, error) {
	def, err := definition.DecodeFile(path)
	if err != nil {
		return nil, fmt.Errorf("maestro load: decode: %w", err)
	}
	if err := validate.WorkflowDefinition(def, validate.Options{}); err != nil {
		return nil, fmt.Errorf("maestro load: validate: %w", err)
	}
	return &Runtime{def: def}, nil
}

// LoadWithValidate is like Load but allows custom validate.Options (schema path, verbose, etc.)
func LoadWithValidate(path string, vopts validate.Options) (*Runtime, error) {
	def, err := definition.DecodeFile(path)
	if err != nil {
		return nil, fmt.Errorf("maestro load: decode: %w", err)
	}
	if err := validate.WorkflowDefinition(def, vopts); err != nil {
		return nil, fmt.Errorf("maestro load: validate: %w", err)
	}
	return &Runtime{def: def}, nil
}

// Compile validates an in-memory definition (for embedded YAML, generated workflows, etc.)
func Compile(def *definition.WorkflowDefinition, vopts validate.Options) (*Runtime, error) {
	if def == nil {
		return nil, fmt.Errorf("maestro compile: nil definition")
	}
	if err := validate.WorkflowDefinition(def, vopts); err != nil {
		return nil, fmt.Errorf("maestro compile: validate: %w", err)
	}
	return &Runtime{def: def}, nil
}

// Definition returns the validated workflow. Do not mutate it.
func (rt *Runtime) Definition() *definition.WorkflowDefinition {
	if rt == nil {
		return nil
	}
	return rt.def
}

// InstanceOptions configures engine.NewInstance for this runtime.
type InstanceOptions struct {
	RunID            string
	InitialVariables map[string]any
	TraceGuards      bool
	ActionRegistry   *engine.Registry
}

// NewInstance creates an engine instance for this runtime's definition.
func (rt *Runtime) NewInstance(opts InstanceOptions) (*engine.Instance, error) {
	if rt == nil || rt.def == nil {
		return nil, engine.ErrNilDefinition
	}
	reg := opts.ActionRegistry
	if reg == nil {
		reg = engine.DefaultRegistry()
	}
	return engine.NewInstance(rt.def, engine.Options{
		RunID:            opts.RunID,
		InitialVariables: opts.InitialVariables,
		TraceGuards:      opts.TraceGuards,
		ActionRegistry:   reg,
	})
}
