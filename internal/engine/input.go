package engine

import (
	"errors"
	"fmt"

	"github.com/justinush/maestro/internal/definition"
)

// SubmitInput applies user input on the current human step.
//
// It validates against inputSchema (when present), merges top-level keys into variables,
// evaluates outbound transitions, runs onExit actions when leaving, and moves currentStepID when a transition fires.
//
// SubmitStayOnStep means input was accepted but no transition matched (caller may prompt again or extend variables).
// SubmitAdvanced means the instance advanced to StepID (now the next step).
// SubmitFailed sets Err (schema failure wraps *InputValidationError when applicable).
func (in *Instance) SubmitInput(input map[string]any) SubmitInputResult {
	if in == nil {
		return SubmitInputResult{Status: SubmitFailed, Err: ErrNilDefinition}
	}

	st, err := in.Step()
	if err != nil {
		return SubmitInputResult{Status: SubmitFailed, StepID: in.CurrentStepID(), Err: err}
	}
	if st.Kind != definition.StepKindHuman {
		return SubmitInputResult{
			Status: SubmitFailed,
			StepID: st.ID,
			Err:    fmt.Errorf("engine: submit input called on non-human step %q (kind %q)", st.ID, st.Kind),
		}
	}

	if !in.onEnterRan {
		if err := in.runOnEnterActions(st); err != nil {
			return SubmitInputResult{Status: SubmitFailed, StepID: st.ID, Err: err}
		}
		in.onEnterRan = true
	}

	if err := in.validateInputSchema(st, input); err != nil {
		return SubmitInputResult{Status: SubmitFailed, StepID: st.ID, Err: err}
	}

	keys := make([]string, 0, len(input))
	for k := range input {
		keys = append(keys, k)
	}
	in.record(Event{Type: EventInputAccepted, StepID: st.ID, InputKeys: keys})

	for k, v := range input {
		if k == "" {
			return SubmitInputResult{
				Status: SubmitFailed,
				StepID: st.ID,
				Err:    errors.New("engine: submit input: empty key"),
			}
		}
		in.variables[k] = v
	}

	next, err := in.pickFirstFiringTransition(st.ID)
	if err != nil {
		if errors.Is(err, ErrNoMatchingTransition) {
			return SubmitInputResult{Status: SubmitStayOnStep, StepID: st.ID}
		}
		return SubmitInputResult{Status: SubmitFailed, StepID: st.ID, Err: err}
	}

	if err := in.runOnExitActions(st); err != nil {
		return SubmitInputResult{Status: SubmitFailed, StepID: st.ID, Err: err}
	}

	in.currentStepID = next
	in.onEnterRan = false
	return SubmitInputResult{Status: SubmitAdvanced, StepID: next}
}
