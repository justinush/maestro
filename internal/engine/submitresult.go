package engine

// SubmitInputStatus is the outcome of SubmitInput.
type SubmitInputStatus int

const (
	// SubmitStayOnStep means input was stored but no outbound transition matched; still on the human step.
	SubmitStayOnStep SubmitInputStatus = iota + 1

	// SubmitAdvanced means input was accepted and currentStepID moved to StepID (the next step).
	SubmitAdvanced

	// SubmitFailed means validation or leaving the step failed; see SubmitInputResult.Err.
	SubmitFailed
)

// SubmitInputResult is returned by Instance.SubmitInput.
//
// StepID is the current step after the call (unchanged on SubmitStayOnStep and SubmitFailed except where noted).
// Err is set only when Status is SubmitFailed.
type SubmitInputResult struct {
	Status SubmitInputStatus
	StepID string
	Err    error
}
