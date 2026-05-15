package validate

import (
	"errors"

	apivalidate "github.com/justinush/maestro/pkg/validate"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var file string
	var opts apivalidate.Options

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate a workflow definition",
		Long: "Decodes a YAML or JSON workflow definition, validates it against JSON Schema (embedded or --schema), " +
			"runs graph checks (including reachability), compiles CEL guards, validates stub params, and compiles each inputSchema as JSON Schema.",
		RunE: func(_ *cobra.Command, _ []string) error {
			if file == "" {
				return errors.New("required flag: --file / -f")
			}
			return apivalidate.WorkflowFile(file, opts)
		},
	}

	cmd.Flags().StringVarP(&file, "file", "f", "", "path to workflow definition (.yaml, .yml, or .json)")
	cmd.Flags().StringVar(&opts.SchemaPath, "schema", "", "optional path to workflow JSON Schema (defaults to embedded v0.1; root URI is schema $id when set)")
	cmd.Flags().BoolVar(&opts.Verbose, "verbose", false, "print richer validation errors (schema, CEL, inputSchema compile)")

	return cmd
}
