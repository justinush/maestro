package engine

import (
	"encoding/json"
	"fmt"

	"github.com/justinush/maestro/internal/definition"
)

func marshalToRawJSON(v any) (definition.RawJSON, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("marshal raw json: %w", err)
	}
	return definition.RawJSON(b), nil
}
