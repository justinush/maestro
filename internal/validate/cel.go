package validate

import (
	"fmt"
	"strings"

	"github.com/justinush/maestro/internal/celenv"
	"github.com/justinush/maestro/internal/definition"
)

// validateCELGuards compiles each non-empty transition when with the shared CEL bindings.
func validateCELGuards(def *definition.WorkflowDefinition, verbose bool) error {
	if def == nil {
		return nil
	}

	env, err := celenv.New()
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
