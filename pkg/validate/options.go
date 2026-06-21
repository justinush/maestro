package validate

import ivalidate "github.com/justinush/maestro/internal/validate"

// Options configures validation (JSON Schema source, verbose diagnostics, etc.).
type Options struct {
	// SchemaPath is the path to a JSON Schema file; empty uses the embedded workflow schema.
	SchemaPath string

	// Verbose includes raw schema validator messages in returned errors.
	Verbose bool

	// AllowedActionTypes lists app-owned action type strings permitted in workflow YAML.
	// Built-in types "stub" and "http" are always valid. Register matching runners on
	// engine.Registry before executing workflows that reference custom types.
	AllowedActionTypes []string
}

func (o Options) toInternal() ivalidate.Options {
	return ivalidate.Options{
		SchemaPath:         o.SchemaPath,
		Verbose:            o.Verbose,
		AllowedActionTypes: append([]string(nil), o.AllowedActionTypes...),
	}
}
