package engine

import (
	"cmp"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/google/cel-go/cel"
)

// pickFirstFiringTransition returns the target step for the first matching transition
// from fromID (sorted by priority ascending, stable on declaration order).
func (in *Instance) pickFirstFiringTransition(fromID string) (string, error) {
	idxs := make([]int, 0)
	for i := range in.def.Transitions {
		if in.def.Transitions[i].From == fromID {
			idxs = append(idxs, i)
		}
	}
	slices.SortStableFunc(idxs, func(a, b int) int {
		pa, pb := in.def.Transitions[a].Priority, in.def.Transitions[b].Priority
		if c := cmp.Compare(pa, pb); c != 0 {
			return c
		}
		return cmp.Compare(a, b)
	})

	for _, ti := range idxs {
		t := in.def.Transitions[ti]
		ok, err := evalWhen(t.When, in.variables)
		if err != nil {
			return "", fmt.Errorf("transition %d from=%q to=%q: %w", ti, t.From, t.To, err)
		}
		if ok {
			return t.To, nil
		}
	}
	return "", ErrNoMatchingTransition
}

func evalWhen(expr string, variables map[string]any) (bool, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return true, nil
	}

	env, err := cel.NewEnv(
		cel.Variable("variables", cel.MapType(cel.StringType, cel.DynType)),
	)
	if err != nil {
		return false, fmt.Errorf("cel env: %w", err)
	}
	ast, iss := env.Compile(expr)
	if iss != nil && iss.Err() != nil {
		return false, fmt.Errorf("cel compile: %w", iss.Err())
	}
	prg, err := env.Program(ast)
	if err != nil {
		return false, fmt.Errorf("cel program: %w", err)
	}

	out, _, err := prg.Eval(map[string]any{"variables": variables})
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrCELGuard, err)
	}

	native, err := out.ConvertToNative(reflect.TypeOf(false))
	if err != nil {
		return false, fmt.Errorf("%w: convert to bool: %w", ErrCELGuard, err)
	}
	b, ok := native.(bool)
	if !ok {
		return false, fmt.Errorf("%w: expected bool, got %T", ErrCELGuard, native)
	}
	return b, nil
}
