package maestro

import (
	"fmt"

	"github.com/justinush/maestro/pkg/definition"
	"github.com/justinush/maestro/pkg/engine"
	"github.com/justinush/maestro/pkg/validate"
)

// Runtime holds a validated workflow ready for NewInstance.
type Runtime struct {
	def *definition.WorkflowDefinition
}

// Load reads and validates a workflow file (default validate options).
func Load(path string) (*Runtime, error) {
	return loadFromPath(path, validate.Options{})
}

// LoadWithValidate is Load with custom validate.Options (custom schema path, verbose errors, etc.).
func LoadWithValidate(path string, vopts validate.Options) (*Runtime, error) {
	return loadFromPath(path, vopts)
}

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

// Compile validates an in-memory definition (embedded YAML, tests, codegen).
func Compile(def *definition.WorkflowDefinition, vopts validate.Options) (*Runtime, error) {
	if def == nil {
		return nil, fmt.Errorf("maestro compile: nil definition")
	}
	if err := validate.WorkflowDefinition(def, vopts); err != nil {
		return nil, fmt.Errorf("maestro compile: validate: %w", err)
	}
	return &Runtime{def: def}, nil
}

// Definition returns the workflow. Do not mutate it.
func (rt *Runtime) Definition() *definition.WorkflowDefinition {
	if rt == nil {
		return nil
	}
	return rt.def
}

// InstanceOptions is passed to NewInstance. Zero value is fine (default stub registry).
type InstanceOptions struct {
	RunID            string
	InitialVariables map[string]any
	TraceGuards      bool
	ActionRegistry   *engine.Registry
}

// NewInstance creates an engine instance for this workflow.
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
