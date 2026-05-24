package definition

import (
	"encoding/json"
	"fmt"

	"gopkg.in/yaml.v3"
)

type RawJSON []byte

func (r *RawJSON) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind == yaml.ScalarNode && (n.Tag == "!!null" || n.Value == "") {
		*r = nil
		return nil
	}
	var v any
	if err := n.Decode(&v); err != nil {
		return err
	}
	if v == nil {
		*r = nil
		return nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("yaml fragment to json: %w", err)
	}
	*r = append(RawJSON(nil), b...)
	return nil
}

func (r *RawJSON) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*r = nil
		return nil
	}
	*r = append(RawJSON(nil), data...)
	return nil
}

func (r RawJSON) MarshalJSON() ([]byte, error) {
	if len(r) == 0 {
		return []byte("null"), nil
	}
	return []byte(r), nil
}

type StepKind string

const (
	StepKindHuman  StepKind = "human"
	StepKindAction StepKind = "action"
	StepKindEnd    StepKind = "end"
)

type Action struct {
	// Type is the action registry key (for example "stub" or "http").
	Type string `json:"type" yaml:"type"`
	// ID is a stable identifier for tracing.
	ID string `json:"id" yaml:"id"`
	// Params is the action-specific JSON object.
	Params RawJSON `json:"params,omitempty" yaml:"params,omitempty"`
}

type Step struct {
	// ID is the unique step identifier within the workflow.
	ID string `json:"id" yaml:"id"`
	// Kind is human, action, or end.
	Kind StepKind `json:"kind" yaml:"kind"`
	// Description is optional documentation for operators.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// Labels are optional searchable tags.
	Labels []string `json:"labels,omitempty" yaml:"labels,omitempty"`
	// Annotations are optional key/value metadata.
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	// PresentationRef is an optional UI hint for human steps.
	PresentationRef string `json:"presentationRef,omitempty" yaml:"presentationRef,omitempty"`
	// InputSchema is optional JSON Schema for human-step SubmitInput payloads.
	InputSchema RawJSON `json:"inputSchema,omitempty" yaml:"inputSchema,omitempty"`
	// OnEnter actions run when the step is entered.
	OnEnter []Action `json:"onEnter,omitempty" yaml:"onEnter,omitempty"`
	// OnExit actions run when leaving the step after a transition fires.
	OnExit []Action `json:"onExit,omitempty" yaml:"onExit,omitempty"`
}

type Transition struct {
	// From is the source step id.
	From string `json:"from" yaml:"from"`
	// To is the destination step id.
	To string `json:"to" yaml:"to"`
	// When is an optional CEL guard; empty means true.
	When string `json:"when,omitempty" yaml:"when,omitempty"`
	// Priority orders guards ascending (lower runs first).
	Priority int `json:"priority,omitempty" yaml:"priority,omitempty"`
}

type WorkflowDefinition struct {
	// SchemaVersion is the workflow schema version (for example "0.1").
	SchemaVersion string `json:"schemaVersion" yaml:"schemaVersion"`
	// ID identifies the workflow across versions.
	ID string `json:"id" yaml:"id"`
	// Version is the workflow revision string stored on RunRecord.
	Version string `json:"version" yaml:"version"`
	// Title is optional display text.
	Title string `json:"title,omitempty" yaml:"title,omitempty"`
	// Description is optional documentation.
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// InitialStepID is where new instances start.
	InitialStepID string `json:"initialStepId" yaml:"initialStepId"`
	// TerminalStepIDs lists end steps that may complete a run.
	TerminalStepIDs []string `json:"terminalStepIds" yaml:"terminalStepIds"`
	// Steps is the workflow graph nodes.
	Steps []Step `json:"steps" yaml:"steps"`
	// Transitions is the workflow graph edges.
	Transitions []Transition `json:"transitions" yaml:"transitions"`
}

func (k *StepKind) UnmarshalYAML(n *yaml.Node) error {
	var s string
	if err := n.Decode(&s); err != nil {
		return err
	}
	switch StepKind(s) {
	case StepKindHuman, StepKindAction, StepKindEnd:
		*k = StepKind(s)
		return nil
	default:
		return fmt.Errorf("invalid step kind %q", s)
	}
}

func (k *StepKind) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	switch StepKind(s) {
	case StepKindHuman, StepKindAction, StepKindEnd:
		*k = StepKind(s)
		return nil
	default:
		return &json.UnmarshalTypeError{Value: s, Type: nil, Field: "kind"}
	}
}
