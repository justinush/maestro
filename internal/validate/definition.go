package validate

import (
	"errors"

	"github.com/justinush/maestro/internal/definition"
)

// WorkflowDefinition validates an already-decoded workflow definition.
// It runs JSON Schema (embedded or opts.SchemaPath), then graph, transition invariants,
// CEL, stub params, http params, step metadata, and inputSchema checks.
func WorkflowDefinition(def *definition.WorkflowDefinition, opts Options) error {
	var schemaErr error
	if opts.SchemaPath != "" {
		schemaErr = validateJSONSchemaFromPath(def, opts.SchemaPath, workflowSchemaURI, opts.Verbose)
	} else {
		schemaErr = validateJSONSchema(def, opts.Verbose)
	}

	return errors.Join(
		schemaErr,
		validateGraph(def),
		validateTransitionInvariants(def),
		validateCELGuards(def, opts.Verbose),
		validateStubActions(def),
		validateHTTPActions(def),
		validateStepMetadata(def),
		validateInputSchemas(def, opts.Verbose),
	)
}
