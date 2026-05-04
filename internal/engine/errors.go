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
)

type UnknownStepError struct {
	StepID string
}

func (e *UnknownStepError) Error() string {
	return fmt.Sprintf("engine: unknown step %q", e.StepID)
}
