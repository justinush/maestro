package validate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/justinushermawan/maestro/internal/definition"
)

type stubParams struct {
	Set map[string]json.RawMessage `json:"set,omitempty"`
}

func decodeStubParams(data []byte) (stubParams, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var p stubParams
	if err := dec.Decode(&p); err != nil {
		return stubParams{}, err
	}
	var tail json.RawMessage
	if err := dec.Decode(&tail); err != nil {
		if err == io.EOF {
			return p, nil
		}
		return stubParams{}, err
	}
	return stubParams{}, fmt.Errorf("trailing JSON after params object")
}

func validateStubActions(def *definition.WorkflowDefinition) error {
	if def == nil {
		return nil
	}
	for i := range def.Steps {
		st := def.Steps[i]
		if err := validateActionListStub(st.ID, "onEnter", st.OnEnter); err != nil {
			return err
		}
		if err := validateActionListStub(st.ID, "onExit", st.OnExit); err != nil {
			return err
		}
	}
	return nil
}

func validateActionListStub(stepID, listName string, actions []definition.Action) error {
	for ai := range actions {
		a := actions[ai]
		if a.Type != "stub" {
			continue
		}
		if len(a.Params) == 0 {
			continue
		}
		p, err := decodeStubParams(a.Params)
		if err != nil {
			return fmt.Errorf(
				"step %q %s[%d] action id=%q type=stub: params: %w",
				stepID, listName, ai, a.ID, err,
			)
		}
		if p.Set == nil {
			continue
		}
		for k, raw := range p.Set {
			if k == "" {
				return fmt.Errorf(
					"step %q %s[%d] action id=%q type=stub: params.set contains empty key",
					stepID, listName, ai, a.ID,
				)
			}
			if !json.Valid(raw) {
				return fmt.Errorf(
					"step %q %s[%d] action id=%q type=stub: params.set[%q] value must be valid JSON",
					stepID, listName, ai, a.ID, k,
				)
			}
		}
	}
	return nil
}
