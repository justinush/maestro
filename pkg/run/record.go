package run

import (
	"github.com/justinush/maestro/pkg/definition"
	"github.com/justinush/maestro/pkg/engine"
)

// RunRecord is the persistence shape for a workflow run (metadata + engine snapshot + revision).
type RunRecord struct {
	RunID           string          `json:"runId"`
	WorkflowID      string          `json:"workflowId"`
	WorkflowVersion string          `json:"workflowVersion"`
	Revision        int64           `json:"revision"`
	State           engine.Snapshot `json:"state"`
}

// RecordFromInstance builds a RunRecord from a live instance.
//
// revision should be 0 before Create; before Save, use the revision returned by Get.
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

// InstanceFromRecord restores an Instance from storage using the same workflow definition as when the run started.
func InstanceFromRecord(rec *RunRecord, def *definition.WorkflowDefinition, opts engine.Options) (*engine.Instance, error) {
	if rec == nil {
		return nil, engine.ErrNilDefinition
	}
	if opts.RunID == "" && rec.RunID != "" {
		opts.RunID = rec.RunID
	}
	return engine.NewInstanceFromSnapshot(def, rec.State, opts)
}
