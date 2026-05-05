package simulate

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

type Scenario struct {
	Workflow         string              `json:"workflow" yaml:"workflow"`
	Validate         *bool               `json:"validate,omitempty" yaml:"validate,omitempty"`
	InitialVariables map[string]any      `json:"initialVariables,omitempty" yaml:"initialVariables,omitempty"`
	Inputs           []ScenarioStepInput `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	MaxSteps         int                 `json:"maxSteps,omitempty" yaml:"maxSteps,omitempty"`
}

type ScenarioStepInput struct {
	StepID string         `json:"stepId" yaml:"stepId"`
	Data   map[string]any `json:"data" yaml:"data"`
}

func (s Scenario) shouldValidate() bool {
	if s.Validate == nil {
		return true
	}
	return *s.Validate
}

func (s Scenario) maxSteps() int {
	if s.MaxSteps <= 0 {
		return 10_000
	}
	return s.MaxSteps
}

func DecodeScenario(path string) (*Scenario, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read scenario: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		dec := json.NewDecoder(bytes.NewReader(b))
		dec.DisallowUnknownFields()
		var sc Scenario
		if err := dec.Decode(&sc); err != nil {
			return nil, fmt.Errorf("parse scenario json: %w", err)
		}
		var tail json.RawMessage
		if err := dec.Decode(&tail); err != nil {
			if err == io.EOF {
				return &sc, nil
			}
			return nil, fmt.Errorf("parse scenario json: %w", err)
		}
		return nil, fmt.Errorf("parse scenario json: trailing content after document")

	case ".yaml", ".yml":
		dec := yaml.NewDecoder(bytes.NewReader(b))
		dec.KnownFields(true)
		var sc Scenario
		if err := dec.Decode(&sc); err != nil {
			return nil, fmt.Errorf("parse scenario yaml: %w", err)
		}
		return &sc, nil

	default:
		return nil, fmt.Errorf("unsupported scenario extension: %q (use .yaml, .yml, or .json)", ext)
	}
}
