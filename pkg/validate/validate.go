// Package validate checks workflow definitions against JSON Schema and Maestro rules.
//
// Use [WorkflowFile] for the same checks as the maestro validate CLI, or
// [WorkflowDefinition] after decoding with [github.com/justinush/maestro/pkg/definition].
// [github.com/justinush/maestro/pkg/maestro].Load* runs validation automatically.
package validate

import (
	ivalidate "github.com/justinush/maestro/internal/validate"
	"github.com/justinush/maestro/pkg/definition"
)

// WorkflowFile loads path and runs the same validation as the maestro validate CLI.
func WorkflowFile(path string, opts Options) error {
	return ivalidate.Workflow(path, opts.toInternal())
}

// WorkflowDefinition validates def in memory (graph, CEL, stubs, input schemas, etc.).
func WorkflowDefinition(def *definition.WorkflowDefinition, opts Options) error {
	return ivalidate.WorkflowDefinition(def, opts.toInternal())
}
