package definition

import idef "github.com/justinush/maestro/internal/definition"

type RawJSON = idef.RawJSON

type StepKind = idef.StepKind

const (
	StepKindHuman  = idef.StepKindHuman
	StepKindAction = idef.StepKindAction
	StepKindEnd    = idef.StepKindEnd
)

type Action = idef.Action
type Step = idef.Step
type Transition = idef.Transition
type WorkflowDefinition = idef.WorkflowDefinition

// DecodeFile reads a .json, .yaml, or .yml workflow file with strict parsing.
func DecodeFile(path string) (*WorkflowDefinition, error) {
	return idef.DecodeFile(path)
}
