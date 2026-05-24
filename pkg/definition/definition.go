// Package definition provides the decoded workflow schema used by Maestro.
//
// Use [DecodeFile], [DecodeYAML], or [DecodeJSON] to load definitions from bytes or disk.
// Application code usually loads through [github.com/justinush/maestro/pkg/maestro] or
// validates with [github.com/justinush/maestro/pkg/validate] before execution.
//
// JSON and YAML decoding are strict: unknown fields are rejected.
package definition

import idef "github.com/justinush/maestro/internal/definition"

// RawJSON holds a JSON object fragment embedded in YAML/JSON workflow files.
type RawJSON = idef.RawJSON

// StepKind is the step discriminator in workflow definitions.
type StepKind = idef.StepKind

// Step kind constants.
const (
	// StepKindHuman is a step that waits for host input via SubmitInput.
	StepKindHuman = idef.StepKindHuman
	// StepKindAction is a step the engine advances automatically (onEnter actions, transitions).
	StepKindAction = idef.StepKindAction
	// StepKindEnd is a terminal step; ids must be listed in WorkflowDefinition.TerminalStepIDs.
	StepKindEnd = idef.StepKindEnd
)

// Action is one onEnter/onExit invocation (type, id, params).
type Action = idef.Action

// Step is a node in the workflow graph (human, action, or end).
type Step = idef.Step

// Transition is a directed edge guarded by an optional CEL expression (when).
type Transition = idef.Transition

// WorkflowDefinition is the root document loaded from workflow YAML/JSON.
type WorkflowDefinition = idef.WorkflowDefinition

// DecodeFile reads a workflow file (.json, .yaml, or .yml). JSON decoding is strict.
func DecodeFile(path string) (*WorkflowDefinition, error) {
	return idef.DecodeFile(path)
}

// DecodeYAML parses workflow YAML bytes (strict: unknown fields rejected).
func DecodeYAML(data []byte) (*WorkflowDefinition, error) {
	return idef.DecodeYAML(data)
}

// DecodeJSON parses workflow JSON bytes (strict: unknown fields and trailing content rejected).
func DecodeJSON(data []byte) (*WorkflowDefinition, error) {
	return idef.DecodeJSON(data)
}
