package definition

import idef "github.com/justinush/maestro/internal/definition"

type RawJSON = idef.RawJSON

type StepKind = idef.StepKind

const (
	StepKindHuman  = idef.StepKindHuman
	StepKindAction = idef.StepKindAction
	StepKindEnd    = idef.StepKindEnd
)

type (
	Action             = idef.Action
	Step               = idef.Step
	Transition         = idef.Transition
	WorkflowDefinition = idef.WorkflowDefinition
)

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
