package run

import "context"

// Store persists RunRecord values. Implementations must be safe for concurrent use.
//
// Revision rules (see RunRecord):
//   - Create: inserts a new run; stored revision becomes 1.
//   - Get: returns a copy including the current revision.
//   - Save: updates only when rec.Revision matches; on success revision increments by 1.
type Store interface {
	// Create inserts a new run. rec.RunID must be set; rec.Revision is stored as 1.
	Create(ctx context.Context, rec *RunRecord) error

	// Get returns a deep copy of the run or ErrNotFound.
	Get(ctx context.Context, runID string) (*RunRecord, error)

	// Save updates the run when rec.Revision matches the stored value (then bumps revision).
	Save(ctx context.Context, rec *RunRecord) error
}
