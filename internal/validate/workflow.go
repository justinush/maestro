package validate

import (
	"errors"

	"github.com/justinush/maestro/internal/definition"
)

// Workflow loads a definition from path and runs schema, graph, CEL, stub, and inputSchema checks.
func Workflow(path string, opts Options) error {
	def, err := definition.DecodeFile(path)
	if err != nil {
		return err
	}

	var schemaErr error
	if opts.SchemaPath != "" {
		schemaErr = validateJSONSchemaFromPath(def, opts.SchemaPath, workflowSchemaURI, opts.Verbose)
	} else {
		schemaErr = validateJSONSchema(def, opts.Verbose)
	}

	return errors.Join(
		schemaErr,
		validateGraph(def),
		validateCELGuards(def, opts.Verbose),
		validateStubActions(def),
		validateInputSchemas(def, opts.Verbose),
	)
}
