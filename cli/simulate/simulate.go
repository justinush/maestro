package simulate

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/justinush/maestro/internal/definition"
	"github.com/justinush/maestro/internal/engine"
	"github.com/justinush/maestro/internal/validate"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	var scenarioPath string
	var trace bool

	cmd := &cobra.Command{
		Use:   "simulate",
		Short: "Simulate a workflow run from a scenario file",
		Long:  "Runs the engine in dry-run mode using a scripted scenario.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if scenarioPath == "" {
				return errors.New("required flag: --scenario / -s")
			}
			return runScenario(cmd, scenarioPath, trace)
		},
	}

	cmd.Flags().StringVarP(&scenarioPath, "scenario", "s", "", "path to simulation scenario file (.yaml, .yml, or .json)")
	cmd.Flags().BoolVar(&trace, "trace", false, "print engine execution trace")
	return cmd
}

func runScenario(cmd *cobra.Command, scenarioPath string, trace bool) error {
	sc, err := DecodeScenario(scenarioPath)
	if err != nil {
		return err
	}
	if sc.Workflow == "" {
		return errors.New("scenario: workflow is required")
	}

	def, err := definition.DecodeFile(sc.Workflow)
	if err != nil {
		return err
	}

	if sc.shouldValidate() {
		if err := validate.WorkflowDefinition(def, validate.Options{}); err != nil {
			return err
		}
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
			if trace {
				printTrace(cmd, in)
			}
			return fmt.Errorf("simulate: stopped without error at step %q", in.CurrentStepID())

		case errors.Is(err, engine.ErrWorkflowCompleted):
			if trace {
				printTrace(cmd, in)
			}
			return checkScenarioAssertionsOnCompletion(cmd, sc, in)

		case errors.Is(err, engine.ErrNeedsInput):
			stepID := in.CurrentStepID()
			if inputIdx >= len(sc.Inputs) {
				if trace {
					printTrace(cmd, in)
				}
				return fmt.Errorf("simulate: step %q needs input but scenario has no more inputs", stepID)
			}

			next := sc.Inputs[inputIdx]
			inputIdx++

			if next.StepID != "" && next.StepID != stepID {
				if trace {
					printTrace(cmd, in)
				}
				return fmt.Errorf("simulate: expected input for step %q, got input for %q", stepID, next.StepID)
			}

			if err := in.SubmitInput(next.Data); err != nil {
				if sc.ExpectErrorContains != "" {
					if trace {
						printTrace(cmd, in)
					}
					return checkScenarioAssertionsOnError(cmd, sc, err)
				}
				if trace {
					printTrace(cmd, in)
				}
				return err
			}
			continue

		default:
			if sc.ExpectErrorContains != "" {
				if trace {
					printTrace(cmd, in)
				}
				return checkScenarioAssertionsOnError(cmd, sc, err)
			}
			if trace {
				printTrace(cmd, in)
			}
			return err
		}
	}

	if trace {
		printTrace(cmd, in)
	}
	return fmt.Errorf("simulate: exceeded maxSteps=%d (possible loop)", stepLimit)
}

func printTrace(cmd *cobra.Command, in *engine.Instance) {
	for _, ev := range in.Events() {
		_, _ = fmt.Fprintln(cmd.OutOrStdout(), ev.String())
	}
}

func checkScenarioAssertionsOnCompletion(cmd *cobra.Command, sc *Scenario, in *engine.Instance) error {
	if sc.ExpectErrorContains != "" {
		return fmt.Errorf("simulate: expected error containing %q, but workflow completed at %q", sc.ExpectErrorContains, in.CurrentStepID())
	}

	if sc.ExpectFinalStep != "" && in.CurrentStepID() != sc.ExpectFinalStep {
		return fmt.Errorf("simulate: expected final step %q, got %q", sc.ExpectFinalStep, in.CurrentStepID())
	}

	if len(sc.ExpectVariables) > 0 {
		vars := in.Variables()
		for k, want := range sc.ExpectVariables {
			got, ok := vars[k]
			if !ok {
				return fmt.Errorf("simulate: expected variable %q to be set", k)
			}
			if !reflect.DeepEqual(got, want) {
				return fmt.Errorf("simulate: variable %q mismatch: want %#v, got %#v", k, want, got)
			}
		}
	}

	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ok (completed %q)\n", in.CurrentStepID())
	return nil
}

func checkScenarioAssertionsOnError(cmd *cobra.Command, sc *Scenario, err error) error {
	if sc.ExpectErrorContains == "" {
		return err
	}
	if !strings.Contains(err.Error(), sc.ExpectErrorContains) {
		return fmt.Errorf("simulate: expected error containing %q, got %q", sc.ExpectErrorContains, err.Error())
	}
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ok (error matched %q)\n", sc.ExpectErrorContains)
	return nil
}
