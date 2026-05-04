package validate

import (
	"errors"
	"fmt"

	"github.com/justinush/maestro/internal/definition"
	"github.com/justinush/maestro/internal/workflowgraph"
)

type transitionKey struct {
	from     string
	to       string
	priority int
}

// validateGraph checks step ids, terminals, transitions, and reachability from initialStepId.
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

	if _, err := workflowgraph.BuildTerminalSet(def.TerminalStepIDs, stepKinds); err != nil {
		return fmt.Errorf("graph: %w", err)
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

// validateReachability ensures every step appears in the BFS closure from initialStepId.
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
