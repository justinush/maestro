package definition

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// DecodeFile reads a workflow definition from disk based on file extension.
//
// JSON: DisallowUnknownFields and reject trailing JSON after the document.
// YAML: KnownFields(true) on the root mapping (unknown keys error).
func DecodeFile(path string) (*WorkflowDefinition, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return DecodeJSON(b)
	case ".yaml", ".yml":
		return DecodeYAML(b)
	default:
		return nil, fmt.Errorf("unsupported file extension: %q (use .yaml, .yml, or .json)", ext)
	}
}

// DecodeYAML parses a single workflow document from YAML bytes (strict: unknown fields rejected).
func DecodeYAML(data []byte) (*WorkflowDefinition, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	var def WorkflowDefinition
	if err := dec.Decode(&def); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}
	return &def, nil
}

// DecodeJSON parses a single workflow document from JSON bytes (strict: unknown fields and trailing content rejected).
func DecodeJSON(data []byte) (*WorkflowDefinition, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var def WorkflowDefinition
	if err := dec.Decode(&def); err != nil {
		return nil, fmt.Errorf("parse json: %w", err)
	}
	var tail json.RawMessage
	if err := dec.Decode(&tail); err != nil {
		if err != io.EOF {
			return nil, fmt.Errorf("parse json: %w", err)
		}
		return &def, nil
	}
	return nil, fmt.Errorf("parse json: trailing content after workflow document")
}
