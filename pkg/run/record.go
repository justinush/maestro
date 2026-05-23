package run

import (
	"github.com/justinush/maestro/pkg/definition"
	"github.com/justinush/maestro/pkg/engine"
)

// RunRecord is the persistence shape for a workflow run.
//
// JSON field names (runId, workflowId, revision, state, ...) are stable for v0.x;
// treat them as part of the public persistence contract.
//
// Revision supports optimistic locking on Store.Save:
//   - Store.Create stores the record with revision 1 (ignore rec.Revision on input).
//   - Store.Get returns the current revision.
//   - Store.Save succeeds only when rec.Revision matches the stored value, then increments it.
//   - run.ErrRevisionConflict when another writer updated the run first.
type RunRecord struct {
	RunID           string          `json:"runId"`
	WorkflowID      string          `json:"workflowId"`
	WorkflowVersion string          `json:"workflowVersion"`
	Revision        int64           `json:"revision"`
	State           engine.Snapshot `json:"state"`
}

// RecordFromInstance builds a RunRecord from a live instance.
//
// Pass revision 0 before Store.Create (Create sets revision to 1).
// Before Store.Save, set revision to the value returned by Store.Get.
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

// InstanceFromRecord restores an engine.Instance from a RunRecord.
//
// Lower-level API: application code should prefer maestro.Runtime.RestoreInstance,
// which binds the record to the validated Runtime workflow definition.
func InstanceFromRecord(rec *RunRecord, def *definition.WorkflowDefinition, opts engine.Options) (*engine.Instance, error) {
	if rec == nil {
		return nil, engine.ErrNilDefinition
	}
	if opts.RunID == "" && rec.RunID != "" {
		opts.RunID = rec.RunID
	}
	return engine.NewInstanceFromSnapshot(def, rec.State, opts)
}
