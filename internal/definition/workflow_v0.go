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
	Type   string  `json:"type" yaml:"type"`
	ID     string  `json:"id" yaml:"id"`
	Params RawJSON `json:"params,omitempty" yaml:"params,omitempty"`
}

type Step struct {
	ID              string            `json:"id" yaml:"id"`
	Kind            StepKind          `json:"kind" yaml:"kind"`
	Description     string            `json:"description,omitempty" yaml:"description,omitempty"`
	Labels          []string          `json:"labels,omitempty" yaml:"labels,omitempty"`
	Annotations     map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	PresentationRef string            `json:"presentationRef,omitempty" yaml:"presentationRef,omitempty"`
	InputSchema     RawJSON           `json:"inputSchema,omitempty" yaml:"inputSchema,omitempty"`
	OnEnter         []Action          `json:"onEnter,omitempty" yaml:"onEnter,omitempty"`
	OnExit          []Action          `json:"onExit,omitempty" yaml:"onExit,omitempty"`
}

type Transition struct {
	From     string `json:"from" yaml:"from"`
	To       string `json:"to" yaml:"to"`
	When     string `json:"when,omitempty" yaml:"when,omitempty"`
	Priority int    `json:"priority,omitempty" yaml:"priority,omitempty"`
}

type WorkflowDefinition struct {
	SchemaVersion   string       `json:"schemaVersion" yaml:"schemaVersion"`
	ID              string       `json:"id" yaml:"id"`
	Version         string       `json:"version" yaml:"version"`
	Title           string       `json:"title,omitempty" yaml:"title,omitempty"`
	Description     string       `json:"description,omitempty" yaml:"description,omitempty"`
	InitialStepID   string       `json:"initialStepId" yaml:"initialStepId"`
	TerminalStepIDs []string     `json:"terminalStepIds" yaml:"terminalStepIds"`
	Steps           []Step       `json:"steps" yaml:"steps"`
	Transitions     []Transition `json:"transitions" yaml:"transitions"`
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
