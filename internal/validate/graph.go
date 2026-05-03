package validate

import (
	"errors"
	"fmt"

	"github.com/justinushermawan/maestro/internal/definition"
)

type transitionKey struct {
	from     string
	to       string
	priority int
}

func validateGraph(def *definition.WorkflowDefinition) error {
	if def == nil {
		return errors.New("graph: definition is nil")
	}
	if len(def.Steps) == 0 {
		return errors.New("graph: steps must be non-empty")
	}

	stepKinds := make(map[string]definition.StepKind, len(def.Steps))
	for i := range def.Steps {
		s := def.Steps[i]
		if s.ID == "" {
			return fmt.Errorf("graph: steps[%d].id is required", i)
		}
		if _, exists := stepKinds[s.ID]; exists {
			return fmt.Errorf("graph: duplicate step id %q", s.ID)
		}
		if s.Kind == "" {
			return fmt.Errorf("graph: steps[%d].kind is required", i)
		}
		stepKinds[s.ID] = s.Kind
	}

	if def.InitialStepID == "" {
		return errors.New("graph: initialStepId must be a non-empty string")
	}
	if _, ok := stepKinds[def.InitialStepID]; !ok {
		return fmt.Errorf("graph: initialStepId %q not found in steps", def.InitialStepID)
	}

	if len(def.TerminalStepIDs) == 0 {
		return errors.New("graph: terminalStepIds must be non-empty")
	}
	termSet := make(map[string]struct{}, len(def.TerminalStepIDs))
	for _, id := range def.TerminalStepIDs {
		if id == "" {
			return errors.New("graph: terminalStepIds entries must be non-empty strings")
		}
		if _, dup := termSet[id]; dup {
			return fmt.Errorf("graph: duplicate terminalStepIds entry %q", id)
		}
		termSet[id] = struct{}{}
		kind, exists := stepKinds[id]
		if !exists {
			return fmt.Errorf("graph: terminal step id %q not found in steps", id)
		}
		if kind != definition.StepKindEnd {
			return fmt.Errorf("graph: terminalStepIds entry %q must have kind \"end\", got %q", id, kind)
		}
	}
	for id, kind := range stepKinds {
		if kind != definition.StepKindEnd {
			continue
		}
		if _, ok := termSet[id]; !ok {
			return fmt.Errorf("graph: step %q has kind \"end\" but is not listed in terminalStepIds", id)
		}
	}

	outgoing := make(map[string]int)
	seenDup := make(map[transitionKey]int)

	for i := range def.Transitions {
		t := def.Transitions[i]
		if t.From == "" || t.To == "" {
			return fmt.Errorf("graph: transitions[%d].from and .to are required", i)
		}
		if _, ok := stepKinds[t.From]; !ok {
			return fmt.Errorf("graph: transitions[%d].from %q not found in steps", i, t.From)
		}
		if _, ok := stepKinds[t.To]; !ok {
			return fmt.Errorf("graph: transitions[%d].to %q not found in steps", i, t.To)
		}
		if stepKinds[t.From] == definition.StepKindEnd {
			return fmt.Errorf("graph: transition from terminal step %q is not allowed (transitions[%d])", t.From, i)
		}
		outgoing[t.From]++

		key := transitionKey{from: t.From, to: t.To, priority: t.Priority}
		if j, ok := seenDup[key]; ok {
			return fmt.Errorf(
				"graph: duplicate transition (from=%q to=%q priority=%d) at transitions[%d] and transitions[%d]",
				t.From, t.To, t.Priority, j, i,
			)
		}
		seenDup[key] = i
	}

	for id, kind := range stepKinds {
		if kind == definition.StepKindEnd {
			continue
		}
		if outgoing[id] < 1 {
			return fmt.Errorf("graph: step %q (kind %q) must have at least one outgoing transition", id, kind)
		}
	}

	if err := validateReachability(def, stepKinds); err != nil {
		return err
	}

	return nil
}

func validateReachability(def *definition.WorkflowDefinition, stepKinds map[string]definition.StepKind) error {
	adj := make(map[string][]string)
	for i := range def.Transitions {
		t := def.Transitions[i]
		adj[t.From] = append(adj[t.From], t.To)
	}

	visited := map[string]struct{}{def.InitialStepID: {}}
	queue := []string{def.InitialStepID}
	for len(queue) > 0 {
		u := queue[0]
		queue = queue[1:]
		for _, v := range adj[u] {
			if _, ok := visited[v]; ok {
				continue
			}
			visited[v] = struct{}{}
			queue = append(queue, v)
		}
	}

	for id := range stepKinds {
		if _, ok := visited[id]; !ok {
			return fmt.Errorf("graph: step %q is not reachable from initialStepId %q", id, def.InitialStepID)
		}
	}
	return nil
}
