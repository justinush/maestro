package maestro

import (
	"fmt"

	"github.com/justinush/maestro/pkg/definition"
	"github.com/justinush/maestro/pkg/engine"
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

// loadFromPath implements Load / LoadWithValidate.
func loadFromPath(path string, vopts validate.Options) (*Runtime, error) {
	def, err := definition.DecodeFile(path)
	if err != nil {
		return nil, fmt.Errorf("maestro load: decode: %w", err)
	}
	if err := validate.WorkflowDefinition(def, vopts); err != nil {
		return nil, fmt.Errorf("maestro load: validate: %w", err)
	}
	return &Runtime{def: def}, nil
}

// Compile validates an in-memory definition (embedded YAML, codegen, tests) without reading a file.
func Compile(def *definition.WorkflowDefinition, vopts validate.Options) (*Runtime, error) {
	if def == nil {
		return nil, fmt.Errorf("maestro compile: nil definition")
	}
	if err := validate.WorkflowDefinition(def, vopts); err != nil {
		return nil, fmt.Errorf("maestro compile: validate: %w", err)
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

// InstanceOptions configures Runtime.NewInstance (registry, trace, variables, run id).
// Zero value uses DefaultRegistry (stub only).
type InstanceOptions struct {
	RunID            string
	InitialVariables map[string]any
	TraceGuards      bool
	ActionRegistry   *engine.Registry
}

// NewInstance builds an engine.Instance for this workflow.
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
