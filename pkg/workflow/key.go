package workflow

import "github.com/justinush/maestro/pkg/definition"

// Key identifies a workflow definition. It matches run.RunRecord workflowId and
// workflowVersion fields.
type Key struct {
	ID      string // WorkflowDefinition.ID
	Version string // WorkflowDefinition.Version
}

// KeyFromDefinition returns the registry key for def.
func KeyFromDefinition(def *definition.WorkflowDefinition) Key {
	if def == nil {
		return Key{}
	}
	return Key{ID: def.ID, Version: def.Version}
}
