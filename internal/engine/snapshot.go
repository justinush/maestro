package engine

import (
	"errors"
	"fmt"

	"github.com/justinush/maestro/internal/definition"
)

// Snapshot is JSON-friendly state for persistence (step, variables, trace).
// CEL programs and input-schema cache are rebuilt on restore, not stored.
type Snapshot struct {
	RunID         string         `json:"runId"`
	TraceGuards   bool           `json:"traceGuards"`
	CurrentStepID string         `json:"currentStepId"`
	OnEnterRan    bool           `json:"onEnterRan"`
	Variables     map[string]any `json:"variables"`
	Events        []Event        `json:"events,omitempty"`
	NextSeq       int            `json:"nextSeq"`
}

// Snapshot copies variables and events for persistence. Nested map values may still be shared.
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

// NewInstanceFromSnapshot rebuilds an instance from a snapshot and the same workflow definition.
// opts.ActionRegistry (or default) must match what you used when the run was created.
// RunID: opts.RunID if non-empty, otherwise snap.RunID.
// TraceGuards: true if opts.TraceGuards or snap.TraceGuards is true.
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
