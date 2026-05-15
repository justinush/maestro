package engine

// SubmitInputStatus is the outcome of SubmitInput.
type SubmitInputStatus int

const (
	// SubmitStayOnStep means input was accepted but no exit transition matched; still on the human step.
	SubmitStayOnStep SubmitInputStatus = iota + 1

	// SubmitAdvanced means input was accepted and the instance moved to the next step.
	SubmitAdvanced

	// SubmitFailed means input was rejected or an action/guard failed while leaving the step.
	SubmitFailed
)

// SubmitInputResult is returned by Instance.SubmitInput.
// StepID is the current step after the call. Err is set only when Status is SubmitFailed.
type SubmitInputResult struct {
	Status SubmitInputStatus
	StepID string
	Err    error
}
