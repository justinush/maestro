package run

import (
	"github.com/justinush/maestro/internal/definition"
	"github.com/justinush/maestro/pkg/engine"
)

// RunRecord is persisted by Store. Revision is for optimistic locking (CAS).
// WorkflowID and WorkflowVersion should match definition.ID and definition.Version used to rebuild the instance.
type RunRecord struct {
	RunID           string          `json:"runId"`
	WorkflowID      string          `json:"workflowId"`
	WorkflowVersion string          `json:"workflowVersion"`
	Revision        int64           `json:"revision"`
	State           engine.Snapshot `json:"state"`
}

// RecordFromInstance builds a RunRecord from a live instance and definition metadata.
// revision should be the last known revision from Store.Get (use 0 when assembling a new record before Create).
func RecordFromInstance(in *engine.Instance, def *definition.WorkflowDefinition, revision int64) *RunRecord {
	if in == nil || def == nil {
		return nil
	}
	s := in.Snapshot()
	return &RunRecord{
		RunID:           in.RunID(),
		WorkflowID:      def.ID,
		WorkflowVersion: def.Version,
		Revision:        revision,
		State:           s,
	}
}

// InstanceFromRecord restores an engine instance from a stored record.
// def must match the workflow used when the run was created (same id/version and graph).
// If opts.RunID is empty, rec.RunID is applied.
func InstanceFromRecord(rec *RunRecord, def *definition.WorkflowDefinition, opts engine.Options) (*engine.Instance, error) {
	if rec == nil {
		return nil, engine.ErrNilDefinition
	}
	if opts.RunID == "" && rec.RunID != "" {
		opts.RunID = rec.RunID
	}
	return engine.NewInstanceFromSnapshot(def, rec.State, opts)
}
