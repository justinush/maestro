package engine

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/justinush/maestro/internal/definition"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

type inputSchemaCache struct {
	mu   sync.Mutex
	byID map[string]*jsonschema.Schema
}

func (c *inputSchemaCache) getOrCompile(step definition.Step) (*jsonschema.Schema, error) {
	if c.byID == nil {
		c.byID = make(map[string]*jsonschema.Schema)
	}
	if sch, ok := c.byID[step.ID]; ok {
		return sch, nil
	}

	raw := strings.TrimSpace(string(step.InputSchema))
	if len(step.InputSchema) == 0 || raw == "" || raw == "null" {
		// No schema => accept any input.
		c.byID[step.ID] = nil
		return nil, nil
	}

	var doc map[string]any
	if err := json.Unmarshal(step.InputSchema, &doc); err != nil {
		return nil, fmt.Errorf("inputSchema step %q: must be a JSON object: %w", step.ID, err)
	}
	if doc == nil {
		return nil, fmt.Errorf("inputSchema step %q: must be a JSON object (got null)", step.ID)
	}

	uri := fmt.Sprintf("urn:maestro:engine:inputSchema:%s", step.ID)
	comp := jsonschema.NewCompiler()
	if err := comp.AddResource(uri, doc); err != nil {
		return nil, fmt.Errorf("inputSchema step %q: load schema resource: %w", step.ID, err)
	}
	sch, err := comp.Compile(uri)
	if err != nil {
		return nil, fmt.Errorf("inputSchema step %q: compile JSON Schema: %w", step.ID, err)
	}

	c.byID[step.ID] = sch
	return sch, nil
}

func (in *Instance) validateInputSchema(step *definition.Step, input map[string]any) error {
	if in == nil {
		return ErrNilDefinition
	}
	if step == nil {
		return fmt.Errorf("engine: step is nil")
	}

	in.inputSchemas.mu.Lock()
	sch, err := in.inputSchemas.getOrCompile(*step)
	in.inputSchemas.mu.Unlock()
	if err != nil {
		return err
	}

	// No schema => accept any input.
	if sch == nil {
		return nil
	}

	if err := sch.Validate(input); err != nil {
		return &InputValidationError{StepID: step.ID, Err: err}
	}
	return nil
}
