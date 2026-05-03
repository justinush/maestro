package validate

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/justinushermawan/maestro/internal/definition"
	"github.com/justinushermawan/maestro/schemas"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

const workflowSchemaURI = "https://github.com/justinushermawan/maestro/schemas/workflow-definition-v0.1.json"

func jsonInstanceForSchema(def *definition.WorkflowDefinition) (map[string]any, error) {
	b, err := json.Marshal(def)
	if err != nil {
		return nil, fmt.Errorf("marshal definition for schema: %w", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("materialize json instance: %w", err)
	}
	return m, nil
}

func schemaRootURI(schemaPath string, schemaDoc map[string]any, embeddedFallback string) string {
	if v, ok := schemaDoc["$id"].(string); ok {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	if strings.TrimSpace(schemaPath) != "" {
		return schemaPath
	}
	return embeddedFallback
}

func formatSchemaErr(context string, verbose bool, err error) error {
	if err == nil {
		return nil
	}
	if !verbose {
		return fmt.Errorf("%s: %w", context, err)
	}
	return fmt.Errorf("%s: %w\n(full) %s", context, err, err.Error())
}

func validateJSONSchema(def *definition.WorkflowDefinition, verbose bool) error {
	var schemaDoc map[string]any
	if err := json.Unmarshal(schemas.WorkflowDefinitionV01, &schemaDoc); err != nil {
		return fmt.Errorf("parse embedded workflow schema: %w", err)
	}

	inst, err := jsonInstanceForSchema(def)
	if err != nil {
		return err
	}

	root := schemaRootURI("", schemaDoc, workflowSchemaURI)

	c := jsonschema.NewCompiler()
	if err := c.AddResource(root, schemaDoc); err != nil {
		return fmt.Errorf("load schema resource: %w", err)
	}
	sch, err := c.Compile(root)
	if err != nil {
		return formatSchemaErr("compile embedded workflow schema", verbose, err)
	}
	if err := sch.Validate(inst); err != nil {
		return formatSchemaErr("schema validation failed", verbose, err)
	}
	return nil
}

func validateJSONSchemaFromPath(def *definition.WorkflowDefinition, schemaPath, embeddedIDFallback string, verbose bool) error {
	b, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("read schema file: %w", err)
	}
	var schemaDoc map[string]any
	if err := json.Unmarshal(b, &schemaDoc); err != nil {
		return fmt.Errorf("parse schema file: %w", err)
	}

	inst, err := jsonInstanceForSchema(def)
	if err != nil {
		return err
	}

	root := schemaRootURI(schemaPath, schemaDoc, embeddedIDFallback)

	c := jsonschema.NewCompiler()
	if err := c.AddResource(root, schemaDoc); err != nil {
		return fmt.Errorf("load schema resource: %w", err)
	}
	sch, err := c.Compile(root)
	if err != nil {
		return formatSchemaErr(fmt.Sprintf("compile schema root %q", root), verbose, err)
	}
	if err := sch.Validate(inst); err != nil {
		return formatSchemaErr("schema validation failed", verbose, err)
	}
	return nil
}
