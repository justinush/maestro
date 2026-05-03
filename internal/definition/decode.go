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

func DecodeFile(path string) (*WorkflowDefinition, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		dec := json.NewDecoder(bytes.NewReader(b))
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

	case ".yaml", ".yml":
		dec := yaml.NewDecoder(bytes.NewReader(b))
		dec.KnownFields(true)
		var def WorkflowDefinition
		if err := dec.Decode(&def); err != nil {
			return nil, fmt.Errorf("parse yaml: %w", err)
		}
		return &def, nil

	default:
		return nil, fmt.Errorf("unsupported file extension: %q (use .yaml, .yml, or .json)", ext)
	}
}
