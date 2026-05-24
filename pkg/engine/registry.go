package engine

import iengine "github.com/justinush/maestro/internal/engine"

// Registry maps workflow action "type" strings to [ActionRunner] implementations.
// Register every type referenced by the workflow before passing the registry to [NewInstance].
type Registry struct {
	impl *iengine.Registry
}

func (r *Registry) inner() *iengine.Registry {
	if r == nil {
		return nil
	}
	return r.impl
}

// Register binds actionType to runner. Duplicate types return an error.
func (r *Registry) Register(actionType string, runner ActionRunner) error {
	return r.inner().Register(actionType, runner)
}

// MustRegister calls [Registry.Register] and panics if Register returns an error.
func (r *Registry) MustRegister(actionType string, runner ActionRunner) {
	r.inner().MustRegister(actionType, runner)
}

// Lookup returns the runner for actionType and whether it exists.
func (r *Registry) Lookup(actionType string) (ActionRunner, bool) {
	return r.inner().Lookup(actionType)
}
