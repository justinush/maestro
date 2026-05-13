package engine

import (
	"fmt"

	"github.com/justinush/maestro/internal/definition"
)

// RunUntilBlocked drives execution forward until it reaches a stop condition.
// It returns ErrNeedsInput when blocked on a human step, ErrWorkflowCompleted when the
// instance reaches a terminal end step, and a non-nil error for any execution failure.
func (in *Instance) RunUntilBlocked() error {
	r := in.RunUntilBlockedResult()
	switch r.Status {
	case RunBlocked:
		return ErrNeedsInput
	case RunCompleted:
		return ErrWorkflowCompleted
	case RunFailed:
		if r.Err != nil {
			return r.Err
		}
		return fmt.Errorf("engine: RunFailed without err")
	default:
		return fmt.Errorf("engine: unexpected run status %v", r.Status)
	}
}

// RunUntilBlockedResult drives execution like RunUntilBlocked but reports normal outcomes
// as RunStatus values instead of sentinel errors. Err is only set when Status == RunFailed.
func (in *Instance) RunUntilBlockedResult() RunResult {
	if in == nil {
		return RunResult{Status: RunFailed, Err: ErrNilDefinition}
	}

	for {
		st, err := in.Step()
		if err != nil {
			return RunResult{Status: RunFailed, StepID: in.CurrentStepID(), Events: in.Events(), Err: err}
		}

		in.record(Event{Type: EventStepEntered, StepID: st.ID})

		if !in.onEnterRan {
			if err := in.runOnEnterActions(st); err != nil {
				return RunResult{Status: RunFailed, StepID: st.ID, Events: in.Events(), Err: err}
			}
			in.onEnterRan = true
		}

		switch st.Kind {
		case definition.StepKindHuman:
			in.record(Event{Type: EventBlocked, StepID: st.ID})
			return RunResult{Status: RunBlocked, StepID: st.ID, Events: in.Events()}

		case definition.StepKindEnd:
			if in.IsTerminal() {
				in.record(Event{Type: EventCompleted, StepID: st.ID})
				return RunResult{Status: RunCompleted, StepID: st.ID, Events: in.Events()}
			}
			return RunResult{
				Status: RunFailed,
				StepID: st.ID,
				Events: in.Events(),
				Err:    fmt.Errorf("engine: step %q has kind end but is not a declared terminal", st.ID),
			}

		case definition.StepKindAction:
			next, err := in.pickFirstFiringTransition(st.ID)
			if err != nil {
				return RunResult{Status: RunFailed, StepID: st.ID, Events: in.Events(), Err: err}
			}
			if err := in.runOnExitActions(st); err != nil {
				return RunResult{Status: RunFailed, StepID: st.ID, Events: in.Events(), Err: err}
			}
			in.currentStepID = next
			in.onEnterRan = false
			continue

		default:
			return RunResult{
				Status: RunFailed,
				StepID: st.ID,
				Events: in.Events(),
				Err:    fmt.Errorf("engine: unsupported step kind %q", st.Kind),
			}
		}
	}
}
