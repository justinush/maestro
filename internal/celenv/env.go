package celenv

import "github.com/google/cel-go/cel"

// New returns the CEL environment for guard expressions.
//
// Bindings:
// - variables: map<string, dyn>.
func New() (*cel.Env, error) {
	return cel.NewEnv(
		cel.Variable("variables", cel.MapType(cel.StringType, cel.DynType)),
	)
}
