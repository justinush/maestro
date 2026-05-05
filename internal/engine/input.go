package engine

import (
	"errors"
	"fmt"

	"github.com/justinush/maestro/internal/definition"
)

// SubmitInput merges input into variables (top-level keys only) and advances out of the current human step.
// It runs onExit for human step, takes the first matching transition, and positions the instance on the next step.
func (in *Instance) SubmitInput(input map[string]any) error {
	if in == nil {
		return ErrNilDefinition
	}

	st, err := in.Step()
	if err != nil {
		return err
	}
	if st.Kind != definition.StepKindHuman {
		return fmt.Errorf("engine: submit input called on non-human step %q (kind %q)", st.ID, st.Kind)
	}

	// If caller skipped RunUntilBlocked, ensure onEnter runs before we accept input.
	if !in.onEnterRan {
		if err := in.runOnEnterActions(st); err != nil {
			return err
		}
		in.onEnterRan = true
	}

	// Shallow merge user input into variables.
	for k, v := range input {
		if k == "" {
			return errors.New("engine: submit input: empty key")
		}
		in.variables[k] = v
	}

	next, err := in.pickFirstFiringTransition(st.ID)
	if err != nil {
		return err
	}

	if err := in.runOnExitActions(st); err != nil {
		return err
	}

	in.currentStepID = next
	in.onEnterRan = false
	return nil
}
