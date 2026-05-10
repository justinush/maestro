package run

import "context"

// Store persists RunRecord values. Implementations must be safe for concurrent use.
type Store interface {
	// Create inserts a new run. rec.RunID must be non-empty. rec.Revision must be 0 (ignored; stored as 1).
	Create(ctx context.Context, rec *RunRecord) error

	// Get returns a deep copy of the run or ErrNotFound.
	Get(ctx context.Context, runID string) (*RunRecord, error)

	// Save updates a run if rec.Revision matches the stored revision; increments revision on success.
	Save(ctx context.Context, rec *RunRecord) error
}
