package engine

import "fmt"

// Registry maps action type strings (JSON/YAML "type") to runners.
// Register all types before passing the registry to NewInstance.
type Registry struct {
	runners map[string]ActionRunner
}

func NewRegistry() *Registry {
	return &Registry{
		runners: make(map[string]ActionRunner),
	}
}

func DefaultRegistry() *Registry {
	r := NewRegistry()
	if err := r.Register("stub", stubRunner{}); err != nil {
		panic("engine: default registry: " + err.Error())
	}
	return r
}

// Register installs a runner for actionType. Duplicate types are rejected.
func (r *Registry) Register(actionType string, runner ActionRunner) error {
	if r == nil {
		return fmt.Errorf("engine: nil registry")
	}
	if actionType == "" {
		return fmt.Errorf("engine: empty action type")
	}
	if runner == nil {
		return fmt.Errorf("engine: nil runner for type %q", actionType)
	}
	if _, exists := r.runners[actionType]; exists {
		return fmt.Errorf("engine: duplicate action type %q", actionType)
	}
	if r.runners == nil {
		r.runners = make(map[string]ActionRunner)
	}
	r.runners[actionType] = runner
	return nil
}

func (r *Registry) MustRegister(actionType string, runner ActionRunner) {
	if err := r.Register(actionType, runner); err != nil {
		panic(err)
	}
}

func (r *Registry) Lookup(actionType string) (ActionRunner, bool) {
	if r == nil || r.runners == nil {
		return nil, false
	}
	runner, ok := r.runners[actionType]
	return runner, ok
}
