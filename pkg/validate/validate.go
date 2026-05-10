package validate

import (
	"github.com/justinush/maestro/internal/definition"
	ivalidate "github.com/justinush/maestro/internal/validate"
)

type Options = ivalidate.Options

// WorkflowFile validates a workflow file on disk (decode + checks).
func WorkflowFile(path string, opts Options) error {
	return ivalidate.Workflow(path, opts)
}

// WorkflowDefinition validates an in-memory definition.
func WorkflowDefinition(def *definition.WorkflowDefinition, opts Options) error {
	return ivalidate.WorkflowDefinition(def, opts)
}
