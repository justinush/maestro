package engine

import (
	"errors"
	"fmt"

	"github.com/justinush/maestro/internal/definition"
)

// SubmitInput validates and merges input into variables (top-level keys only).
//
// Return values:
// - advanced=true,  err=nil: input accepted and an exit transition fired; instance moved to next step.
// - advanced=false, err=nil: input accepted but no exit transition matched; instance stays on the human step.
// - err!=nil: input rejected (schema) or execution failure (CEL eval, onExit, etc).
func (in *Instance) SubmitInput(input map[string]any) (bool, error) {
	if in == nil {
		return false, ErrNilDefinition
	}

	st, err := in.Step()
	if err != nil {
		return false, err
	}
	if st.Kind != definition.StepKindHuman {
		return false, fmt.Errorf("engine: submit input called on non-human step %q (kind %q)", st.ID, st.Kind)
	}

	// If caller skipped RunUntilBlocked, ensure onEnter runs before we accept input.
	if !in.onEnterRan {
		if err := in.runOnEnterActions(st); err != nil {
			return false, err
		}
		in.onEnterRan = true
	}

	// Enforce inputSchema before mutation.
	if err := in.validateInputSchema(st, input); err != nil {
		return false, err
	}

	keys := make([]string, 0, len(input))
	for k := range input {
		keys = append(keys, k)
	}
	in.record(Event{Type: EventInputAccepted, StepID: st.ID, InputKeys: keys})

	// Shallow merge user input into variables.
	for k, v := range input {
		if k == "" {
			return false, errors.New("engine: submit input: empty key")
		}
		in.variables[k] = v
	}

	next, err := in.pickFirstFiringTransition(st.ID)
	if err != nil {
		// Special case: no match means "stay on the human step" (input accepted).
		if errors.Is(err, ErrNoMatchingTransition) {
			return false, nil
		}
		return false, err
	}

	// Only run onExit when we actually leave the step.
	if err := in.runOnExitActions(st); err != nil {
		return false, err
	}

	in.currentStepID = next
	in.onEnterRan = false
	return true, nil
}
