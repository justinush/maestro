package validate

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/justinush/maestro/internal/definition"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// validateInputSchemas compiles each non-empty step inputSchema as a standalone JSON Schema root.
func validateInputSchemas(def *definition.WorkflowDefinition, verbose bool) error {
	if def == nil {
		return nil
	}
	for i := range def.Steps {
		st := def.Steps[i]
		raw := strings.TrimSpace(string(st.InputSchema))
		if len(st.InputSchema) == 0 || raw == "" || raw == "null" {
			continue
		}
		var doc map[string]any
		if err := json.Unmarshal(st.InputSchema, &doc); err != nil {
			return fmt.Errorf("inputSchema step %q: must be a JSON object: %w", st.ID, err)
		}
		if doc == nil {
			return fmt.Errorf("inputSchema step %q: must be a JSON object (got null)", st.ID)
		}

		uri := fmt.Sprintf("urn:maestro:validate:inputSchema:%s", st.ID)
		c := jsonschema.NewCompiler()
		if err := c.AddResource(uri, doc); err != nil {
			return fmt.Errorf("inputSchema step %q: load schema resource: %w", st.ID, err)
		}
		if _, err := c.Compile(uri); err != nil {
			if verbose {
				return fmt.Errorf("inputSchema step %q: compile JSON Schema: %w\n(full) %s", st.ID, err, err.Error())
			}
			return fmt.Errorf("inputSchema step %q: compile JSON Schema: %w", st.ID, err)
		}
	}
	return nil
}
