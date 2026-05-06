package engine

import (
	"fmt"

	"github.com/justinush/maestro/internal/definition"
	"github.com/justinush/maestro/internal/stub"
)

func (in *Instance) runOnEnterActions(st *definition.Step) error {
	return in.runActionList(st.ID, "onEnter", st.OnEnter)
}

func (in *Instance) runOnExitActions(st *definition.Step) error {
	return in.runActionList(st.ID, "onExit", st.OnExit)
}

func (in *Instance) runActionList(stepID, listName string, actions []definition.Action) error {
	for i := range actions {
		a := actions[i]
		switch a.Type {
		case "stub":
			if len(a.Params) == 0 {
				in.record(Event{
					Type:       EventActionRan,
					StepID:     stepID,
					ActionList: listName,
					ActionID:   a.ID,
					ActionType: a.Type,
				})
				continue
			}
			p, err := stub.DecodeParams(a.Params)
			if err != nil {
				return fmt.Errorf("step %q %s[%d] action %q: %w", stepID, listName, i, a.ID, err)
			}
			if err := applyStubSet(in.variables, p); err != nil {
				return fmt.Errorf("step %q %s[%d] action %q: %w", stepID, listName, i, a.ID, err)
			}
			in.record(Event{
				Type:       EventActionRan,
				StepID:     stepID,
				ActionList: listName,
				ActionID:   a.ID,
				ActionType: a.Type,
			})
		default:
			return fmt.Errorf("%w: step %q %s[%d] action %q type %q", ErrUnknownActionType, stepID, listName, i, a.ID, a.Type)
		}
	}
	return nil
}
