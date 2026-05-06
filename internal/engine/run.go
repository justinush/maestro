package engine

import (
	"fmt"

	"github.com/justinush/maestro/internal/definition"
)

// RunUntilBlocked drives execution forward until it reaches a stop condition.
// It returns ErrNeedsInput when blocked on a human step, ErrWorkflowCompleted when the
// instance reaches a terminal end step, and a non-nil error for any execution failure.
func (in *Instance) RunUntilBlocked() error {
	if in == nil {
		return ErrNilDefinition
	}

	for {
		st, err := in.Step()
		if err != nil {
			return err
		}

		in.record(Event{Type: EventStepEntered, StepID: st.ID})

		if !in.onEnterRan {
			if err := in.runOnEnterActions(st); err != nil {
				return err
			}
			in.onEnterRan = true
		}

		switch st.Kind {
		case definition.StepKindHuman:
			in.record(Event{Type: EventBlocked, StepID: st.ID})
			return ErrNeedsInput

		case definition.StepKindEnd:
			if in.IsTerminal() {
				in.record(Event{Type: EventCompleted, StepID: st.ID})
				return ErrWorkflowCompleted
			}
			return fmt.Errorf("engine: step %q has kind end but is not a declared terminal", st.ID)

		case definition.StepKindAction:
			next, err := in.pickFirstFiringTransition(st.ID)
			if err != nil {
				return err
			}
			if err := in.runOnExitActions(st); err != nil {
				return err
			}
			in.currentStepID = next
			in.onEnterRan = false
			continue

		default:
			return fmt.Errorf("engine: unsupported step kind %q", st.Kind)
		}
	}
}
