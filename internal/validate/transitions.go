package validate

import (
	"fmt"
	"strings"

	"github.com/justinush/maestro/internal/definition"
)

type fromPriority struct {
	from     string
	priority int
}

func validateTransitionInvariants(def *definition.WorkflowDefinition) error {
	if def == nil {
		return nil
	}

	// Track unconditional transitions by (from, priority).
	seenUnconditional := make(map[fromPriority]int)

	for i := range def.Transitions {
		t := def.Transitions[i]
		when := strings.TrimSpace(t.When)
		if when != "" {
			continue
		}
		k := fromPriority{from: t.From, priority: t.Priority}
		if j, ok := seenUnconditional[k]; ok {
			return fmt.Errorf(
				"transitions: ambiguous unconditional transitions from=%q priority=%d at transitions[%d] and transitions[%d]",
				t.From, t.Priority, j, i,
			)
		}
		seenUnconditional[k] = i
	}

	return nil
}
