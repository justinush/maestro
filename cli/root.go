package cli

import (
	"github.com/justinush/maestro/cli/simulate"
	"github.com/justinush/maestro/cli/validate"
	"github.com/spf13/cobra"
)

var Version = "dev"

func Execute() error {
	return newRootCmd().Execute()
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "maestro",
		Short:         "CLI tool for Maestro",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       Version,
	}

	root.SetVersionTemplate("{{.Version}}\n")

	root.AddCommand(validate.NewCommand())
	root.AddCommand(simulate.NewCommand())

	return root
}
