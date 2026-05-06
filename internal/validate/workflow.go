package validate

import (
	"github.com/justinush/maestro/internal/definition"
)

// Workflow loads a definition from path and validates it (schema + graph + CEL + stub + inputSchemas).
func Workflow(path string, opts Options) error {
	def, err := definition.DecodeFile(path)
	if err != nil {
		return err
	}
	return WorkflowDefinition(def, opts)
}
