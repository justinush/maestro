package engine

import iengine "github.com/justinush/maestro/internal/engine"

// Registry maps workflow action "type" strings to [ActionRunner] implementations.
// Register every type referenced by the workflow before passing to registry to [NewInstance].
type Registry struct {
	*iengine.Registry
}
