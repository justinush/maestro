package engine

// RunStatus describes why RunUntilBlocked stopped.
type RunStatus int

const (
	// RunBlocked means the instance is waiting on a human step.
	RunBlocked RunStatus = iota + 1

	// RunCompleted means the instance reached a declared terminal end step.
	RunCompleted

	// RunFailed means execution failed.
	RunFailed
)

// RunResult is the outcome of Instance.RunUntilBlocked.
// For RunBlocked and RunCompleted, Err is nil.
type RunResult struct {
	Status RunStatus
	StepID string
	Events []Event
	Err    error
}
