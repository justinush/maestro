package engine

import (
	"fmt"
	"maps"

	"github.com/google/cel-go/cel"
	"github.com/justinush/maestro/internal/definition"
	"github.com/justinush/maestro/internal/workflowgraph"
)

// Options configures a new workflow instance (identity, initial variables, tracing, action runners).
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

// newInstanceShell validates def, indexes steps and terminals, and wires the action registry.
// It returns an Instance without an initial step position or variables; callers must set those.
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

// NewInstance creates an Instance positioned at def.InitialStepID with a shallow copy of opts.InitialVariables.
// It validates step ids, initial step presence, and terminal rules (via workflowgraph.BuildTerminalSet).
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

// Definition returns the workflow bound to this instance, or nil if in is nil.
// Callers must not mutate the returned pointer or its contents.
func (in *Instance) Definition() *definition.WorkflowDefinition {
	if in == nil {
		return nil
	}
	return in.def
}

// RunID returns Options.RunID (may be empty).
func (in *Instance) RunID() string {
	if in == nil {
		return ""
	}
	return in.runID
}

// CurrentStepID returns the step the engine is currently on.
func (in *Instance) CurrentStepID() string {
	if in == nil {
		return ""
	}
	return in.currentStepID
}

// Step returns the definition for the current step.
func (in *Instance) Step() (*definition.Step, error) {
	if in == nil {
		return nil, ErrNilDefinition
	}
	return in.StepByID(in.currentStepID)
}

// StepByID returns the step definition for id.
// If id is not in the workflow, the error satisfies errors.As(..., *UnknownStepError).
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

// IsTerminal reports whether the current step is both kind "end" and listed in terminalStepIds.
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

// stepKindsFromSteps builds id->kind for terminal validation.
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

// Events returns a copy of recorded trace events (safe to read without racing the instance).
func (in *Instance) Events() []Event {
	if in == nil {
		return nil
	}
	return append([]Event(nil), in.events...)
}

// record appends an event and assigns monotonic Seq and RunID.
func (in *Instance) record(ev Event) {
	in.nextSeq++
	ev.Seq = in.nextSeq
	ev.RunID = in.runID
	in.events = append(in.events, ev)
}
