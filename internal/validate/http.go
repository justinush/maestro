package validate

import (
	"fmt"

	"github.com/justinush/maestro/internal/definition"
	"github.com/justinush/maestro/internal/httpaction"
)

// validateHTTPActions checks http params on every step's onEnter/onExit lists.
func validateHTTPActions(def *definition.WorkflowDefinition) error {
	if def == nil {
		return nil
	}
	for i := range def.Steps {
		st := def.Steps[i]
		if err := validateActionListHTTP(st.ID, "onEnter", st.OnEnter); err != nil {
			return err
		}
		if err := validateActionListHTTP(st.ID, "onExit", st.OnExit); err != nil {
			return err
		}
	}
	return nil
}

func validateActionListHTTP(stepID, listName string, actions []definition.Action) error {
	for ai := range actions {
		a := actions[ai]
		if a.Type != "http" {
			continue
		}
		if len(a.Params) == 0 {
			return fmt.Errorf(
				"step %q %s[%d] action id=%q type=http: params are required",
				stepID, listName, ai, a.ID,
			)
		}
		if _, err := httpaction.DecodeParams(a.Params); err != nil {
			return fmt.Errorf(
				"step %q %s[%d] action id=%q type=http: params: %w",
				stepID, listName, ai, a.ID, err,
			)
		}
	}
	return nil
}
