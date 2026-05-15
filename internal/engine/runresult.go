package engine

// RunStatus is why RunUntilBlocked stopped.
type RunStatus int

const (
	// RunBlocked - waiting on a human step; call SubmitInput next.
	RunBlocked RunStatus = iota + 1

	// RunCompleted - on a declared terminal end step.
	RunCompleted

	// RunFailed - execution error; see RunResult.Err.
	RunFailed
)

// RunResult is returned by RunUntilBlocked.
// Err is non-nil only when Status is RunFailed.
type RunResult struct {
	Status RunStatus
	StepID string
	Events []Event
	Err    error
}
