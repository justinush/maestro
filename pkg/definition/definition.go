// Package definition provides the decoded workflow schema used by Maestro.
//
// Use [DecodeFile], [DecodeYAML], or [DecodeJSON] to load definitions from bytes or disk.
// Application code usually loads through [github.com/justinush/maestro/pkg/maestro] or
// validates with [github.com/justinush/maestro/pkg/validate] before execution.
//
// JSON and YAML decoding are strict: unknown fields are rejected.
package definition

import idef "github.com/justinush/maestro/internal/definition"

// RawJSON holds a JSON object fragment embedded in YAML/JSON workflow files
// (for example action params or inputSchema).
type RawJSON = idef.RawJSON

// StepKind is the step discriminator in workflow definitions (human, action, or end).
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

// Action is one onEnter/onExit invocation
//
// Fields:
//   - Type: registry key (for example "stub" or "http")
//   - ID: stable identifier for tracing
//   - Params: action-specific JSON object (RawJSON)
type Action = idef.Action

// Step is a node in the workflow graph.
//
// Fields:
//   - ID, Kind (required)
//   - Description, Labels, Annotations, PresentationRef (optional metadata / UI)
//   - InputSchema: optional JSON Schema for human-step SubmitInput payloads
//   - OnEnter, OnExit: action lists run when entering or leaving the step
type Step = idef.Step

// Transition is a directed edge between steps.
//
// Fields:
//   - From, To: step ids
//   - When: optional CEL guard (empty means true)
//   - Priority: ascending order when multiple guards are evaluated (lower first)
type Transition = idef.Transition

// WorkflowDefinition is the root document loaded from workflow YAML/JSON.
//
// Fields:
//   - SchemaVersion: schema version (for example "0.1")
//   - ID, Version: workflow identity (Version is stored on run.RunRecord)
//   - Title, Description: optional documentation
//   - InitialStepID: where new instances start
//   - TerminalStepIDs: end step ids that may complete a run
//   - Steps, Transitions: the workflow graph
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
