package engine

import (
	"fmt"
	"net/http"
)

// Registry maps workflow action "type" strings to ActionRunner implementations.
// Register everything needed by the workflow before passing the registry to NewInstance.
type Registry struct {
	runners map[string]ActionRunner
}

// NewRegistry returns an empty registry (no built-in types registered).
func NewRegistry() *Registry {
	return &Registry{
		runners: make(map[string]ActionRunner),
	}
}

// DefaultRegistry returns a registry with the built-in "stub" runner registered.
func DefaultRegistry() *Registry {
	r := NewRegistry()
	if err := r.Register("stub", stubRunner{}); err != nil {
		panic("engine: default registry: " + err.Error())
	}
	return r
}

// RegistryWithHTTP returns DefaultRegistry plus an "http" runner using client.
func RegistryWithHTTP(client *http.Client) *Registry {
	r := DefaultRegistry()
	if err := r.Register("http", NewHTTPRunner(client)); err != nil {
		panic("engine: registry with http: " + err.Error())
	}
	return r
}

// Register binds actionType to runner. Duplicate types return an error.
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

// MustRegister calls Register and panics if Register returns an error.
func (r *Registry) MustRegister(actionType string, runner ActionRunner) {
	if err := r.Register(actionType, runner); err != nil {
		panic(err)
	}
}

// Lookup returns the runner for actionType and whether it exists.
func (r *Registry) Lookup(actionType string) (ActionRunner, bool) {
	if r == nil || r.runners == nil {
		return nil, false
	}
	runner, ok := r.runners[actionType]
	return runner, ok
}
