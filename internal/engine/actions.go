package engine

import (
	"fmt"

	"github.com/justinush/maestro/internal/definition"
)

// runOnEnterActions runs st.OnEnter in order.
func (in *Instance) runOnEnterActions(st *definition.Step) error {
	return in.runActionList(st.ID, "onEnter", st.OnEnter)
}

// runOnExitActions runs st.OnExit in order.
func (in *Instance) runOnExitActions(st *definition.Step) error {
	return in.runActionList(st.ID, "onExit", st.OnExit)
}

// runActionList dispatches each action to the registry and records EventActionRan on success.
func (in *Instance) runActionList(stepID, listName string, actions []definition.Action) error {
	for i := range actions {
		a := actions[i]
		runner, ok := in.actionReg.Lookup(a.Type)
		if !ok {
			return fmt.Errorf("%w: step %q %s[%d] action %q type %q", ErrUnknownActionType, stepID, listName, i, a.ID, a.Type)
		}
		ctx := ActionContext{
			RunID:     in.runID,
			StepID:    stepID,
			ListName:  listName,
			Index:     i,
			Action:    a,
			Variables: in.variables,
		}
		if err := runner.Run(ctx); err != nil {
			return fmt.Errorf("step %q %s[%d] action %q: %w", stepID, listName, i, a.ID, err)
		}
		in.record(Event{
			Type:       EventActionRan,
			StepID:     stepID,
			ActionList: listName,
			ActionID:   a.ID,
			ActionType: a.Type,
		})
	}
	return nil
}
