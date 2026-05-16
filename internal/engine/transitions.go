package engine

import (
	"cmp"
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/justinush/maestro/internal/celenv"
)

// pickFirstFiringTransition evaluates outbound transitions from fromID in priority order
// (ascending priority, stable tie-break by declaration index) and returns the first whose When guard passes.
// ErrNoMatchingTransition means every guard evaluated to false (not an evaluation error).
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
		ok, err := in.evalWhenForTransition(ti, in.variables)
		if in.traceGuards {
			ev := Event{
				Type:            EventTransitionGuard,
				TransitionIndex: ti,
				FromStepID:      t.From,
				ToStepID:        t.To,
			}

			switch {
			case err != nil:
				ev.GuardResult = "error"
				ev.GuardError = err.Error()
			case ok:
				ev.GuardResult = "true"
			default:
				ev.GuardResult = "false"
			}

			in.record(ev)
		}

		if err != nil {
			return "", fmt.Errorf("transition %d from=%q to=%q: %w", ti, t.From, t.To, err)
		}
		if ok {
			in.record(Event{
				Type:            EventTransitionTaken,
				TransitionIndex: ti,
				FromStepID:      t.From,
				ToStepID:        t.To,
			})
			return t.To, nil
		}
	}
	return "", ErrNoMatchingTransition
}

// evalWhenForTransition evaluates transitions[ti].When against variables ("variables" activation).
// Empty When means true. Errors from CEL evaluation wrap ErrCELGuard.
func (in *Instance) evalWhenForTransition(ti int, variables map[string]any) (bool, error) {
	if in == nil {
		return false, ErrNilDefinition
	}
	expr := strings.TrimSpace(in.def.Transitions[ti].When)
	if expr == "" {
		return true, nil
	}

	prg, err := in.celProgramForTransition(ti)
	if err != nil {
		return false, err
	}

	out, _, err := prg.Eval(map[string]any{"variables": variables})
	if err != nil {
		return false, fmt.Errorf("%w: %w", ErrCELGuard, err)
	}

	native, err := out.ConvertToNative(reflect.TypeFor[bool]())
	if err != nil {
		return false, fmt.Errorf("%w: convert to bool: %w", ErrCELGuard, err)
	}
	b, ok := native.(bool)
	if !ok {
		return false, fmt.Errorf("%w: expected bool, got %T", ErrCELGuard, native)
	}
	return b, nil
}

// celProgramForTransition returns a compiled CEL program for transitions[ti].When, cached per transition index.
func (in *Instance) celProgramForTransition(ti int) (cel.Program, error) {
	if in.celPrograms != nil {
		if prg, ok := in.celPrograms[ti]; ok {
			return prg, nil
		}
	}

	expr := strings.TrimSpace(in.def.Transitions[ti].When)
	if expr == "" {
		return nil, fmt.Errorf("engine: celProgramForTransition called with empty when at index %d", ti)
	}

	env, err := celenv.New()
	if err != nil {
		return nil, fmt.Errorf("cel env: %w", err)
	}

	ast, iss := env.Compile(expr)
	if iss != nil && iss.Err() != nil {
		return nil, fmt.Errorf("cel compile: %w", iss.Err())
	}

	prg, err := env.Program(ast)
	if err != nil {
		return nil, fmt.Errorf("cel program: %w", err)
	}

	if in.celPrograms == nil {
		in.celPrograms = make(map[int]cel.Program)
	}
	in.celPrograms[ti] = prg
	return prg, nil
}
