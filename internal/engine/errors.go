package engine

import (
	"errors"
	"fmt"
)

var (
	ErrNilDefinition = errors.New("engine: workflow definition is nil")

	ErrEmptyInitialStepID = errors.New("engine: initialStepId is empty")

	ErrInitialStepUnknown = errors.New("engine: initialStepId not found in steps")

	// ErrEmptyStepID is returned when a step has an empty id.
	ErrEmptyStepID = errors.New("engine: step id is empty")

	// ErrEmptyStepKind is returned when a step has an empty kind.
	ErrEmptyStepKind = errors.New("engine: step kind is empty")

	// ErrDuplicateStepID is returned when two steps share the same id.
	ErrDuplicateStepID = errors.New("engine: duplicate step id")

	// ErrNeedsInput is returned when execution stops on a human step waiting for input.
	ErrNeedsInput = errors.New("engine: human step needs input")

	// ErrWorkflowCompleted is returned when execution reaches a declared terminal end step.
	ErrWorkflowCompleted = errors.New("engine: workflow completed")

	// ErrNoMatchingTransition is returned when no transition guard from an action step matches.
	ErrNoMatchingTransition = errors.New("engine: no matching transition")

	// ErrUnknownActionType is returned for action types other than stub.
	ErrUnknownActionType = errors.New("engine: unknown action type")

	// ErrCELGuard is returned when a when expression fails at runtime (not compile-time).
	ErrCELGuard = errors.New("engine: cel guard evaluation failed")
)

type UnknownStepError struct {
	StepID string
}

func (e *UnknownStepError) Error() string {
	return fmt.Sprintf("engine: unknown step %q", e.StepID)
}
