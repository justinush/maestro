package workflow

import (
	"fmt"
	"sort"
	"sync"

	"github.com/justinush/maestro/pkg/engine"
	"github.com/justinush/maestro/pkg/maestro"
	"github.com/justinush/maestro/pkg/run"
)

// Registry maps workflow identity (Key) to a validated maestro.Runtime.
//
// A Registry is safe for concurrent Lookup, NewInstance, and RestoreInstance after
// all workflows are registered. Do not call Register concurrently with those methods.
type Registry struct {
	mu    sync.RWMutex
	byKey map[Key]*maestro.Runtime
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{
		byKey: make(map[Key]*maestro.Runtime),
	}
}

// Register adds rt indexed by its WorkflowDefinition ID and Version.
// Returns ErrDuplicateKey when the same Key is already registered.
func (r *Registry) Register(rt *maestro.Runtime) error {
	if r == nil {
		return fmt.Errorf("workflow: nil registry")
	}
	if rt == nil {
		return fmt.Errorf("workflow: nil runtime")
	}
	def := rt.Definition()
	if def == nil {
		return fmt.Errorf("workflow: runtime has nil definition")
	}
	key := KeyFromDefinition(def)
	if key.ID == "" {
		return fmt.Errorf("workflow: workflow definition has empty id")
	}
	if key.Version == "" {
		return fmt.Errorf("workflow: workflow definition %q has empty version", key.ID)
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.byKey[key]; exists {
		return fmt.Errorf(
			"%w: workflow id %q version %q",
			ErrDuplicateKey,
			key.ID,
			key.Version,
		)
	}
	r.byKey[key] = rt
	return nil
}

// Lookup returns the runtime for key, or ErrNotFound.
func (r *Registry) Lookup(key Key) (*maestro.Runtime, error) {
	if r == nil {
		return nil, fmt.Errorf("workflow: nil registry")
	}
	r.mu.RLock()
	rt, ok := r.byKey[key]
	r.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf(
			"%w: workflow id %q version %q",
			ErrNotFound,
			key.ID,
			key.Version,
		)
	}
	return rt, nil
}

// NewInstance looks up key and starts a new engine.Instance.
func (r *Registry) NewInstance(key Key, opts maestro.InstanceOptions) (*engine.Instance, error) {
	rt, err := r.Lookup(key)
	if err != nil {
		return nil, err
	}
	return rt.NewInstance(opts)
}

// RestoreInstance looks up the workflow for rec.WorkflowID and rec.WorkflowVersion,
// verifies the loaded definition matches the record, then restores the instance.
func (r *Registry) RestoreInstance(rec *run.RunRecord, opts maestro.InstanceOptions) (*engine.Instance, error) {
	if r == nil {
		return nil, fmt.Errorf("workflow: nil registry")
	}
	if rec == nil {
		return nil, fmt.Errorf("workflow: nil run record")
	}
	key := Key{ID: rec.WorkflowID, Version: rec.WorkflowVersion}
	rt, err := r.Lookup(key)
	if err != nil {
		return nil, err
	}
	def := rt.Definition()
	if err := checkRecordDefinitionMatch(rec, def); err != nil {
		return nil, err
	}
	return rt.RestoreInstance(rec, opts)
}

// Contains reports whether key is registered.
func (r *Registry) Contains(key Key) bool {
	if r == nil {
		return false
	}
	r.mu.RLock()
	_, ok := r.byKey[key]
	r.mu.RUnlock()
	return ok
}

// List returns all registered keys in stable sorted order (id, then version).
func (r *Registry) List() []Key {
	if r == nil {
		return nil
	}
	r.mu.RLock()
	keys := make([]Key, 0, len(r.byKey))
	for k := range r.byKey {
		keys = append(keys, k)
	}
	r.mu.RUnlock()
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].ID != keys[j].ID {
			return keys[i].ID < keys[j].ID
		}
		return keys[i].Version < keys[j].Version
	})
	return keys
}
