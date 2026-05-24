package engine

import (
	iengine "github.com/justinush/maestro/internal/engine"
	"github.com/justinush/maestro/pkg/definition"
)

// Instance is a single workflow run. Create one with [NewInstance] or [NewInstanceFromSnapshot].
//
// Advance execution with [Instance.RunUntilBlocked] and [Instance.SubmitInput].
// Inspect [RunResult] and [SubmitInputResult] status values before reading Err.
type Instance struct {
	impl *iengine.Instance
}

func (in *Instance) inner() *iengine.Instance {
	if in == nil {
		return nil
	}
	return in.impl
}

// Definition returns the workflow bound to this instance, or nil if in is nil.
// Callers must not mutate the returned pointer or its contents.
func (in *Instance) Definition() *definition.WorkflowDefinition {
	return in.inner().Definition()
}

// RunID returns the correlation id for this run (may be empty).
func (in *Instance) RunID() string {
	return in.inner().RunID()
}

// CurrentStepID returns the step the engine is currently on.
func (in *Instance) CurrentStepID() string {
	return in.inner().CurrentStepID()
}

// Step returns the definition for the current step.
func (in *Instance) Step() (*definition.Step, error) {
	return in.inner().Step()
}

// StepByID returns the step definition for id.
// If id is not in the workflow, the error satisfies [AsUnknownStepError].
func (in *Instance) StepByID(id string) (*definition.Step, error) {
	return in.inner().StepByID(id)
}

// Variables returns a shallow copy of the variable map (nested values are still shared).
func (in *Instance) Variables() map[string]any {
	return in.inner().Variables()
}

// IsTerminal reports whether the current step is both kind "end" and listed in terminalStepIds.
func (in *Instance) IsTerminal() bool {
	return in.inner().IsTerminal()
}

// Events returns a copy of recorded trace events (safe to read without racing the instance).
func (in *Instance) Events() []Event {
	return in.inner().Events()
}

// RunUntilBlocked advances execution until the engine must stop and yield to the host.
//
// [RunBlocked]: current step is human - call [Instance.SubmitInput] next.
// [RunCompleted]: reached a declared terminal end step.
// [RunFailed]: Err describes the failure; StepID and Events reflect progress up to the failure.
//
// Returned Events is a snapshot at return time (same ordering as [Instance.Events]).
func (in *Instance) RunUntilBlocked() RunResult {
	return in.inner().RunUntilBlocked()
}

// SubmitInput applies user input on the current human step.
//
// It validates against inputSchema (when present), merges top-level keys into variables,
// evaluates outbound transitions, runs onExit actions when leaving, and moves currentStepID when a transition fires.
//
// [SubmitStayOnStep] means input was accepted but no transition matched (caller may prompt again).
// [SubmitAdvanced] means the instance advanced to SubmitInputResult.StepID.
// [SubmitFailed] sets Err (schema failure wraps [InputValidationError] when applicable).
func (in *Instance) SubmitInput(input map[string]any) SubmitInputResult {
	return in.inner().SubmitInput(input)
}

// Snapshot builds a [Snapshot] value from the current instance.
// Variables and Events are shallow-copied; nested values may still alias live state.
func (in *Instance) Snapshot() Snapshot {
	return in.inner().Snapshot()
}
