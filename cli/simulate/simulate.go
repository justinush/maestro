package simulate

import (
	"errors"
	"fmt"

	"github.com/justinush/maestro/internal/definition"
	"github.com/justinush/maestro/internal/engine"
	"github.com/justinush/maestro/internal/validate"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var scenarioPath string

	cmd := &cobra.Command{
		Use:   "simulate",
		Short: "Simulate a workflow run from a scenario file",
		Long:  "Runs the engine in dry-run mode using a scripted scenario.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if scenarioPath == "" {
				return errors.New("required flag: --scenario / -s")
			}
			return runScenario(cmd, scenarioPath)
		},
	}

	cmd.Flags().StringVarP(&scenarioPath, "scenario", "s", "", "path to simulation scenario file (.yaml, .yml, or .json)")
	return cmd
}

func runScenario(cmd *cobra.Command, scenarioPath string) error {
	sc, err := DecodeScenario(scenarioPath)
	if err != nil {
		return err
	}
	if sc.Workflow == "" {
		return errors.New("scenario: workflow is required")
	}

	if sc.shouldValidate() {
		if err := validate.Workflow(sc.Workflow, validate.Options{}); err != nil {
			return err
		}
	}

	def, err := definition.DecodeFile(sc.Workflow)
	if err != nil {
		return err
	}

	in, err := engine.NewInstance(def, engine.Options{
		InitialVariables: sc.InitialVariables,
	})
	if err != nil {
		return err
	}

	inputIdx := 0
	stepLimit := sc.maxSteps()

	for range stepLimit {
		err := in.RunUntilBlocked()
		switch {
		case err == nil:
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "simulate: stopped without error at step %q\n", in.CurrentStepID())
			return nil

		case errors.Is(err, engine.ErrWorkflowCompleted):
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "completed at %q\n", in.CurrentStepID())
			return nil

		case errors.Is(err, engine.ErrNeedsInput):
			stepID := in.CurrentStepID()
			if inputIdx >= len(sc.Inputs) {
				return fmt.Errorf("simulate: step %q needs input but scenario has no more inputs", stepID)
			}
			next := sc.Inputs[inputIdx]
			inputIdx++

			if next.StepID != "" && next.StepID != stepID {
				return fmt.Errorf("simulate: expected input for step %q, got input for %q", stepID, next.StepID)
			}

			if err := in.SubmitInput(next.Data); err != nil {
				return err
			}
			continue

		default:
			return err
		}
	}

	return fmt.Errorf("simulate: exceeded maxSteps=%d (possible loop)", stepLimit)
}
