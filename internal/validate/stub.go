package validate

import (
	"encoding/json"
	"fmt"

	"github.com/justinush/maestro/internal/definition"
	"github.com/justinush/maestro/internal/stub"
)

// validateStubActions checks stub params and set values on every step's onEnter and onExit lists.
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

// validateActionListStub validates stub actions in one onEnter/onExit list.
func validateActionListStub(stepID, listName string, actions []definition.Action) error {
	for ai := range actions {
		a := actions[ai]
		if a.Type != "stub" {
			continue
		}
		if len(a.Params) == 0 {
			continue
		}
		p, err := stub.DecodeParams(a.Params)
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
