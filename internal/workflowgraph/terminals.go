package workflowgraph

import (
	"errors"
	"fmt"

	"github.com/justinush/maestro/internal/definition"
)

var (
	// ErrNoTerminalSteps is returned when terminalStepIDs is empty.
	ErrNoTerminalSteps = errors.New("terminalStepIds must be non-empty")

	// ErrEmptyTerminalStepID is returned when a terminalStepIds entry is empty.
	ErrEmptyTerminalStepID = errors.New("terminalStepIds entry is empty")

	// ErrDuplicateTerminalStepID is returned when terminalStepIds repeats an id.
	ErrDuplicateTerminalStepID = errors.New("duplicate terminalStepIds entry")

	// ErrTerminalStepUnknown is returned when a terminal id is not a step id.
	ErrTerminalStepUnknown = errors.New("terminal step id not found in steps")

	// ErrTerminalStepNotEnd is returned when a terminal id does not have kind end.
	ErrTerminalStepNotEnd = errors.New("terminalStepIds entry must have kind end")

	// ErrEndStepNotTerminal is returned when a step has kind end but is not listed in terminalStepIds.
	ErrEndStepNotTerminal = errors.New("step has kind end but is not listed in terminalStepIds")
)

// BuildTerminalSet checks terminalStepIDs against step kinds and returns the terminal id set.
// stepKinds must include every step in the workflow (id->kind).
func BuildTerminalSet(terminalStepIDs []string, stepKinds map[string]definition.StepKind) (map[string]struct{}, error) {
	if len(terminalStepIDs) == 0 {
		return nil, ErrNoTerminalSteps
	}
	term := make(map[string]struct{}, len(terminalStepIDs))
	for i, id := range terminalStepIDs {
		if id == "" {
			return nil, fmt.Errorf("%w at terminalStepIds[%d]", ErrEmptyTerminalStepID, i)
		}
		if _, dup := term[id]; dup {
			return nil, fmt.Errorf("%w: %q", ErrDuplicateTerminalStepID, id)
		}
		term[id] = struct{}{}
		kind, exists := stepKinds[id]
		if !exists {
			return nil, fmt.Errorf("%w: %q", ErrTerminalStepUnknown, id)
		}
		if kind != definition.StepKindEnd {
			return nil, fmt.Errorf("%w: %q has kind %q", ErrTerminalStepNotEnd, id, kind)
		}
	}
	for id, kind := range stepKinds {
		if kind != definition.StepKindEnd {
			continue
		}
		if _, ok := term[id]; !ok {
			return nil, fmt.Errorf("%w: %q", ErrEndStepNotTerminal, id)
		}
	}
	return term, nil
}
