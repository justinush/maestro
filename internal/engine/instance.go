package engine

import (
	"fmt"
	"maps"

	"github.com/google/cel-go/cel"
	"github.com/justinush/maestro/internal/definition"
	"github.com/justinush/maestro/internal/workflowgraph"
)

type Options struct {
	RunID string

	InitialVariables map[string]any

	TraceGuards bool

	ActionRegistry *Registry
}

type Instance struct {
	def         *definition.WorkflowDefinition
	runID       string
	traceGuards bool
	stepsByID   map[string]definition.Step
	terminal    map[string]struct{}
	actionReg   *Registry

	// Execution state
	currentStepID string
	variables     map[string]any
	onEnterRan    bool

	// Lazy caches
	inputSchemas inputSchemaCache
	celPrograms  map[int]cel.Program

	// Trace
	events  []Event
	nextSeq int
}

// newInstanceShell validates def, builds step/terminal maps, and attaches registry.
// It does not set currentStepID, variables, onEnterRan, events, or nextSeq.
func newInstanceShell(def *definition.WorkflowDefinition, opts Options) (*Instance, error) {
	if def == nil {
		return nil, ErrNilDefinition
	}
	if def.InitialStepID == "" {
		return nil, ErrEmptyInitialStepID
	}

	stepsByID := make(map[string]definition.Step, len(def.Steps))
	for i := range def.Steps {
		s := def.Steps[i]
		if s.ID == "" {
			return nil, fmt.Errorf("%w at steps[%d]", ErrEmptyStepID, i)
		}
		if s.Kind == "" {
			return nil, fmt.Errorf("%w at steps[%d]", ErrEmptyStepKind, i)
		}
		if _, exists := stepsByID[s.ID]; exists {
			return nil, fmt.Errorf("%w: %q", ErrDuplicateStepID, s.ID)
		}
		stepsByID[s.ID] = s
	}
	if _, ok := stepsByID[def.InitialStepID]; !ok {
		return nil, fmt.Errorf("%w: %q", ErrInitialStepUnknown, def.InitialStepID)
	}

	stepKinds := stepKindsFromSteps(stepsByID)
	term, err := workflowgraph.BuildTerminalSet(def.TerminalStepIDs, stepKinds)
	if err != nil {
		return nil, fmt.Errorf("engine: terminals: %w", err)
	}

	reg := opts.ActionRegistry
	if reg == nil {
		reg = DefaultRegistry()
	}

	return &Instance{
		def:           def,
		runID:         opts.RunID,
		traceGuards:   opts.TraceGuards,
		stepsByID:     stepsByID,
		terminal:      term,
		actionReg:     reg,
		currentStepID: "",
		variables:     nil,
		onEnterRan:    false,
		celPrograms:   nil,
		events:        nil,
		nextSeq:       0,
	}, nil
}

// NewInstance creates an instance at initialStepId with a copy of opts.InitialVariables.
// It checks step ids, initial step presence, and terminal rules via workflowgraph.BuildTerminalSet.
func NewInstance(def *definition.WorkflowDefinition, opts Options) (*Instance, error) {
	in, err := newInstanceShell(def, opts)
	if err != nil {
		return nil, err
	}

	vars := make(map[string]any)
	maps.Copy(vars, opts.InitialVariables)

	in.variables = vars
	in.currentStepID = def.InitialStepID
	in.onEnterRan = false
	in.events = nil
	in.nextSeq = 0
	return in, nil
}

// Definition returns the workflow for this instance, or nil when the receiver is nil.
// Do not mutate the returned value.
func (in *Instance) Definition() *definition.WorkflowDefinition {
	if in == nil {
		return nil
	}
	return in.def
}

// RunID returns the correlation id from Options, or empty if unset.
func (in *Instance) RunID() string {
	if in == nil {
		return ""
	}
	return in.runID
}

// CurrentStepID returns the step the instance is positioned on.
func (in *Instance) CurrentStepID() string {
	if in == nil {
		return ""
	}
	return in.currentStepID
}

// Step loads the current step definition, or an error if the id is unknown.
func (in *Instance) Step() (*definition.Step, error) {
	if in == nil {
		return nil, ErrNilDefinition
	}
	return in.StepByID(in.currentStepID)
}

// StepByID loads a step by id; returns UnknownStepError when id is missing.
func (in *Instance) StepByID(id string) (*definition.Step, error) {
	if in == nil {
		return nil, ErrNilDefinition
	}
	s, ok := in.stepsByID[id]
	if !ok {
		return nil, &UnknownStepError{StepID: id}
	}
	cp := s
	return &cp, nil
}

// Variables returns a shallow copy of the variable map (nested values are still shared).
func (in *Instance) Variables() map[string]any {
	if in == nil {
		return nil
	}
	return shallowCopyMap(in.variables)
}

// IsTerminal reports whether the current step is listed as terminal and has kind end.
func (in *Instance) IsTerminal() bool {
	if in == nil {
		return false
	}
	if _, ok := in.terminal[in.currentStepID]; !ok {
		return false
	}
	s, ok := in.stepsByID[in.currentStepID]
	if !ok {
		return false
	}
	return s.Kind == definition.StepKindEnd
}

// stepKindsFromSteps builds id→kind for terminal validation.
func stepKindsFromSteps(stepsByID map[string]definition.Step) map[string]definition.StepKind {
	m := make(map[string]definition.StepKind, len(stepsByID))
	for id, s := range stepsByID {
		m[id] = s.Kind
	}
	return m
}

func shallowCopyMap(m map[string]any) map[string]any {
	if m == nil {
		return map[string]any{}
	}
	out := make(map[string]any, len(m))
	maps.Copy(out, m)
	return out
}

// Events returns a snapshot of the execution trace so far.
func (in *Instance) Events() []Event {
	if in == nil {
		return nil
	}
	return append([]Event(nil), in.events...)
}

func (in *Instance) record(ev Event) {
	in.nextSeq++
	ev.Seq = in.nextSeq
	ev.RunID = in.runID
	in.events = append(in.events, ev)
}
