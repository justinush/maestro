package engine

// RunStatus explains why RunUntilBlocked stopped.
type RunStatus int

const (
	// RunBlocked means execution paused on a human step; call SubmitInput next.
	RunBlocked RunStatus = iota + 1

	// RunCompleted means execution reached a declared terminal end step.
	RunCompleted

	// RunFailed means execution stopped with an error; see RunResult.Err.
	RunFailed
)

// RunResult is the outcome of RunUntilBlocked.
//
// Events is a point-in-time snapshot (matches Instance.Events() at return).
// Err is non-nil only when Status is RunFailed.
type RunResult struct {
	Status RunStatus
	StepID string
	Events []Event
	Err    error
}
