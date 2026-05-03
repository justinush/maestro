package simulate

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var scenario string

	cmd := &cobra.Command{
		Use:   "simulate",
		Short: "Simulate a workflow run from a scenario file",
		Long:  "Runs the engine in dry-run mode using a scripted scenario.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if scenario == "" {
				return errors.New("required flag: --scenario / -s")
			}
			_, err := fmt.Fprintf(cmd.OutOrStdout(), "simulate %q (TODO)\n", scenario)
			return err
		},
	}

	cmd.Flags().StringVarP(&scenario, "scenario", "s", "", "path to simulation scenario file")

	return cmd
}
