package engine

import "github.com/justinush/maestro/internal/definition"

type ActionContext struct {
	RunID     string
	StepID    string
	ListName  string // "onEnter" or "onExit"
	Index     int
	Action    definition.Action
	Variables map[string]any
}

// ActionRunner executes a single workflow action.
type ActionRunner interface {
	Run(ctx ActionContext) error
}
