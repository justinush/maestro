package engine

import (
	"errors"
	"fmt"

	"github.com/justinush/maestro/internal/definition"
)

// Snapshot is JSON-friendly runtime state: correlation id, step position, variables, and trace.
// It does not store compiled CEL programs or compiled input schemas; those are rebuilt after restore.
type Snapshot struct {
	RunID         string         `json:"runId"`
	TraceGuards   bool           `json:"traceGuards"`
	CurrentStepID string         `json:"currentStepId"`
	OnEnterRan    bool           `json:"onEnterRan"`
	Variables     map[string]any `json:"variables"`
	Events        []Event        `json:"events,omitempty"`
	NextSeq       int            `json:"nextSeq"`
}

// Snapshot builds a Snapshot value from the current instance.
// Variables and Events are shallow-copied (top-level maps/slices only); nested values may still alias live state.
func (in *Instance) Snapshot() Snapshot {
	if in == nil {
		return Snapshot{}
	}
	return Snapshot{
		RunID:         in.runID,
		TraceGuards:   in.traceGuards,
		CurrentStepID: in.currentStepID,
		OnEnterRan:    in.onEnterRan,
		Variables:     shallowCopyMap(in.variables),
		Events:        append([]Event(nil), in.events...),
		NextSeq:       in.nextSeq,
	}
}

// NewInstanceFromSnapshot restores an Instance from Snapshot plus the workflow used when the run began.
// def must describe the same graph as at snapshot time (same ids and compatible steps/transitions).
//
// Registry: pass the same ActionRegistry semantics as then the run was created (or rely on defaults consistently).
// RunID uses opts.RunID when non-empty, otherwise snap.RunID.
// TraceGuards is true if either opts.TraceGuards or snap.TraceGuards is true.
func NewInstanceFromSnapshot(def *definition.WorkflowDefinition, snap Snapshot, opts Options) (*Instance, error) {
	if snap.CurrentStepID == "" {
		return nil, errors.New("engine: snapshot currentStepId is empty")
	}
	in, err := newInstanceShell(def, opts)
	if err != nil {
		return nil, err
	}
	if _, ok := in.stepsByID[snap.CurrentStepID]; !ok {
		return nil, fmt.Errorf("%w: snapshot step %q", ErrInitialStepUnknown, snap.CurrentStepID)
	}

	in.currentStepID = snap.CurrentStepID
	in.onEnterRan = snap.OnEnterRan
	in.variables = shallowCopyMap(snap.Variables)
	in.events = append([]Event(nil), snap.Events...)
	in.nextSeq = snap.NextSeq

	if opts.RunID != "" {
		in.runID = opts.RunID
	} else {
		in.runID = snap.RunID
	}
	in.traceGuards = opts.TraceGuards || snap.TraceGuards
	return in, nil
}
