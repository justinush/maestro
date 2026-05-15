package engine

import (
	"errors"
	"fmt"
)

var (
	ErrNilDefinition = errors.New("engine: workflow definition is nil")

	ErrEmptyInitialStepID = errors.New("engine: initialStepId is empty")

	ErrInitialStepUnknown = errors.New("engine: initialStepId not found in steps")

	ErrEmptyStepID = errors.New("engine: step id is empty")

	ErrEmptyStepKind = errors.New("engine: step kind is empty")

	ErrDuplicateStepID = errors.New("engine: duplicate step id")

	ErrNoMatchingTransition = errors.New("engine: no matching transition")

	ErrUnknownActionType = errors.New("engine: unknown action type")

	ErrCELGuard = errors.New("engine: cel guard evaluation failed")
)

type UnknownStepError struct {
	StepID string
}

func (e *UnknownStepError) Error() string {
	return fmt.Sprintf("engine: unknown step %q", e.StepID)
}
