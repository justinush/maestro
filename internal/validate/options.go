package validate

// Options configures Workflow and WorkflowDefinition validation.
type Options struct {
	// SchemaPath is the path to a JSON Schema file; empty uses the embedded workflow schema.
	SchemaPath string

	// Verbose includes raw schema validator messages in returned errors.
	Verbose bool

	// AllowedActionTypes lists app-owned action type strings that may appear in workflow YAML.
	// Built-in types "stub" and "http" are always valid.
	// Register a matching runner on engine.Registry before execution.
	AllowedActionTypes []string
}
