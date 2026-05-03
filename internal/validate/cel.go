package validate

import (
	"fmt"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/justinushermawan/maestro/internal/definition"
)

func validateCELGuards(def *definition.WorkflowDefinition, verbose bool) error {
	if def == nil {
		return nil
	}
	env, err := cel.NewEnv(
		cel.Variable("variables", cel.MapType(cel.StringType, cel.DynType)),
	)
	if err != nil {
		return fmt.Errorf("cel: create environment: %w", err)
	}
	for i := range def.Transitions {
		expr := strings.TrimSpace(def.Transitions[i].When)
		if expr == "" {
			continue
		}
		_, iss := env.Compile(expr)
		if iss != nil && iss.Err() != nil {
			wrapped := iss.Err()
			if verbose {
				return fmt.Errorf("cel: transitions[%d].when compile failed: %w\n(details) %s", i, wrapped, iss.String())
			}
			return fmt.Errorf("cel: transitions[%d].when compile failed: %w", i, wrapped)
		}
	}
	return nil
}
