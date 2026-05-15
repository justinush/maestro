package run

import (
	"github.com/justinush/maestro/pkg/definition"
	"github.com/justinush/maestro/pkg/engine"
)

// RunRecord is what Store persists. Revision supports optimistic locking on Save.
type RunRecord struct {
	RunID           string          `json:"runId"`
	WorkflowID      string          `json:"workflowId"`
	WorkflowVersion string          `json:"workflowVersion"`
	Revision        int64           `json:"revision"`
	State           engine.Snapshot `json:"state"`
}

// RecordFromInstance builds a record from a live instance.
// revision: use 0 before Create; use the revision from Get before Save.
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

// InstanceFromRecord restores an instance. def must match the workflow used when the run started.
func InstanceFromRecord(rec *RunRecord, def *definition.WorkflowDefinition, opts engine.Options) (*engine.Instance, error) {
	if rec == nil {
		return nil, engine.ErrNilDefinition
	}
	if opts.RunID == "" && rec.RunID != "" {
		opts.RunID = rec.RunID
	}
	return engine.NewInstanceFromSnapshot(def, rec.State, opts)
}
