package engine

import iengine "github.com/justinush/maestro/internal/engine"

// Instance is a single workflow run. Create one with [NewInstance] or [NewInstanceFromSnapshot].
//
// Advance execution with [Instance.RunUntilBlocked] and [Instance.SubmitInput].
// Inspect [RunResult] and [SubmitInputResult] status values before reading Err.
type Instance struct {
	*iengine.Instance
}
