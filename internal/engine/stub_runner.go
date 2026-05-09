package engine

import (
	"encoding/json"
	"fmt"

	"github.com/justinush/maestro/internal/stub"
)

type stubRunner struct{}

// NewStubRunner returns the built-in stub runner (params.set → variables).
func NewStubRunner() ActionRunner {
	return stubRunner{}
}

func (stubRunner) Run(ctx ActionContext) error {
	if len(ctx.Action.Params) == 0 {
		return nil
	}
	p, err := stub.DecodeParams(ctx.Action.Params)
	if err != nil {
		return err
	}
	return applyStubSet(ctx.Variables, p)
}

// applyStubSet merges stub params.set into vars (top-level keys only; each value replaces the key).
func applyStubSet(vars map[string]any, p stub.Params) error {
	if p.Set == nil {
		return nil
	}
	for k, raw := range p.Set {
		if k == "" {
			return fmt.Errorf("stub params.set: empty key")
		}
		var v any
		if err := json.Unmarshal(raw, &v); err != nil {
			return fmt.Errorf("stub params.set[%q]: %w", k, err)
		}
		vars[k] = v
	}
	return nil
}
