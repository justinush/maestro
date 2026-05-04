package engine

import (
	"encoding/json"
	"fmt"

	"github.com/justinush/maestro/internal/stub"
)

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
