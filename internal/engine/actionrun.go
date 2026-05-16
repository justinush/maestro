package engine

import "github.com/justinush/maestro/internal/definition"

// ActionContext is the input to ActionRunner.Run for one action invocation.
type ActionContext struct {
	RunID     string
	StepID    string
	ListName  string // "onEnter" or "onExit"
	Index     int
	Action    definition.Action
	Variables map[string]any
}

// ActionRunner executes a single workflow action (stub, http, or custom types).
type ActionRunner interface {
	Run(ctx ActionContext) error
}
