package engine

import "fmt"

// InputValidationError means SubmitInput input failed JSON Schema validation for a human step.
type InputValidationError struct {
	StepID string
	Err    error
}

func (e *InputValidationError) Error() string {
	return fmt.Sprintf("engine: inputSchema validation failed for step %q: %v", e.StepID, e.Err)
}

func (e *InputValidationError) Unwrap() error {
	return e.Err
}
