package validate

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/justinush/maestro/internal/definition"
	"github.com/justinush/maestro/schemas"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// workflowSchemaURI is the canonical $id for the embedded workflow-definition v0.1 schema.
const workflowSchemaURI = "https://github.com/justinush/maestro/schemas/workflow-definition-v0.1.json"

// jsonInstanceForSchema builds the JSON object jsonschema validates against.
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

// schemaRootURI picks the compiler root; schema $id, else the schema file path, else embeddedFallback.
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

// formatSchemaErr wraps a schema error; verbose appends the raw library message.
func formatSchemaErr(context string, verbose bool, err error) error {
	if err == nil {
		return nil
	}
	if !verbose {
		return fmt.Errorf("%s: %w", context, err)
	}
	return fmt.Errorf("%s: %w\n(full) %s", context, err, err.Error())
}

func loadSchemaDocFromBytes(b []byte, opts Options) (map[string]any, error) {
	var schemaDoc map[string]any
	if err := json.Unmarshal(b, &schemaDoc); err != nil {
		return nil, fmt.Errorf("parse workflow schema: %w", err)
	}
	if err := extendActionTypeEnum(schemaDoc, opts.AllowedActionTypes); err != nil {
		return nil, err
	}
	return schemaDoc, nil
}

// validateJSONSchema validates def against the embedded v0.1 workflow schema.
func validateJSONSchema(def *definition.WorkflowDefinition, opts Options) error {
	schemaDoc, err := loadSchemaDocFromBytes(schemas.WorkflowDefinitionV01, opts)
	if err != nil {
		return err
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
		return formatSchemaErr("compile embedded workflow schema", opts.Verbose, err)
	}
	if err := sch.Validate(inst); err != nil {
		return formatSchemaErr("schema validation failed", opts.Verbose, err)
	}
	return nil
}

// validateJSONSchemaFromPath validates def against the JSON Schema file at schemaPath.
func validateJSONSchemaFromPath(def *definition.WorkflowDefinition, schemaPath, embeddedIDFallback string, opts Options) error {
	b, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("read schema file: %w", err)
	}
	schemaDoc, err := loadSchemaDocFromBytes(b, opts)
	if err != nil {
		return err
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
		return formatSchemaErr(fmt.Sprintf("compile schema root %q", root), opts.Verbose, err)
	}
	if err := sch.Validate(inst); err != nil {
		return formatSchemaErr("schema validation failed", opts.Verbose, err)
	}
	return nil
}
