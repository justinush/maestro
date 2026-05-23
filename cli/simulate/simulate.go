package simulate

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/justinush/maestro/pkg/definition"
	"github.com/justinush/maestro/pkg/engine"
	"github.com/justinush/maestro/pkg/validate"
	"github.com/spf13/cobra"
)

type traceFormat string

const (
	traceText traceFormat = "text"
	traceJSON traceFormat = "json"
)

func NewCommand() *cobra.Command {
	var scenarioPath string
	var trace bool
	var traceGuards bool
	var format string

	cmd := &cobra.Command{
		Use:   "simulate",
		Short: "Simulate a workflow run from a scenario file",
		Long:  "Runs the engine in dry-run mode using a scripted scenario.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if scenarioPath == "" {
				return errors.New("required flag: --scenario / -s")
			}
			if traceGuards && !trace {
				trace = true
			}
			tf := traceFormat(strings.ToLower(strings.TrimSpace(format)))
			if tf == "" {
				tf = traceText
			}
			if tf != traceText && tf != traceJSON {
				return fmt.Errorf("invalid --trace-format %q (use %q or %q)", format, traceText, traceJSON)
			}

			return runScenario(cmd, scenarioPath, trace, traceGuards, tf)
		},
	}

	cmd.Flags().StringVarP(&scenarioPath, "scenario", "s", "", "path to simulation scenario file (.yaml, .yml, or .json)")
	cmd.Flags().BoolVar(&trace, "trace", false, "print engine execution trace")
	cmd.Flags().BoolVar(&traceGuards, "trace-guards", false, "record and print transition guard results (enables --trace)")
	cmd.Flags().StringVar(&format, "trace-format", "text", "trace format: text or json (only used when --trace is set)")
	return cmd
}

func runScenario(cmd *cobra.Command, scenarioPath string, trace, traceGuards bool, tf traceFormat) error {
	sc, err := DecodeScenario(scenarioPath)
	if err != nil {
		return err
	}
	if sc.Workflow == "" {
		return errors.New("scenario: workflow is required")
	}

	def, err := definition.DecodeFile(sc.Workflow)
	if err != nil {
		if sc.ExpectErrorContains != "" {
			return checkScenarioAssertionsOnError(cmd, sc, err)
		}
		return err
	}

	if sc.shouldValidate() {
		if err := validate.WorkflowDefinition(def, validate.Options{}); err != nil {
			if sc.ExpectErrorContains != "" {
				return checkScenarioAssertionsOnError(cmd, sc, err)
			}
			return err
		}
	}

	in, err := engine.NewInstance(def, engine.Options{
		InitialVariables: sc.InitialVariables,
		TraceGuards:      traceGuards,
		ActionRegistry:   engine.RegistryWithHTTP(simulateHTTPClient()),
	})
	if err != nil {
		if sc.ExpectErrorContains != "" {
			return checkScenarioAssertionsOnError(cmd, sc, err)
		}
		return err
	}

	maybeTrace := func() error {
		if !trace {
			return nil
		}
		return printTrace(cmd, in, tf)
	}

	inputIdx := 0
	stepLimit := sc.maxSteps()

	for range stepLimit {
		res := in.RunUntilBlocked()
		switch res.Status {
		case engine.RunCompleted:
			if err := maybeTrace(); err != nil {
				return err
			}
			return checkScenarioAssertionsOnCompletion(cmd, sc, in)

		case engine.RunBlocked:
			stepID := res.StepID
			if inputIdx >= len(sc.Inputs) {
				if err := maybeTrace(); err != nil {
					return err
				}
				return fmt.Errorf("simulate: step %q needs input but scenario has no more inputs", stepID)
			}

			next := sc.Inputs[inputIdx]
			inputIdx++

			if next.StepID != "" && next.StepID != stepID {
				if err := maybeTrace(); err != nil {
					return err
				}
				return fmt.Errorf("simulate: expected input for step %q, got input for %q", stepID, next.StepID)
			}

			sub := in.SubmitInput(next.Data)
			switch sub.Status {
			case engine.SubmitAdvanced:
				// continue outer loop
			case engine.SubmitStayOnStep:
				if trace {
					_, _ = fmt.Fprintf(
						cmd.ErrOrStderr(),
						"simulate: input accepted at %q but no transition matched; staying on step\n",
						stepID,
					)
				}
			case engine.SubmitFailed:
				err := sub.Err
				if sc.ExpectErrorContains != "" {
					if traceErr := maybeTrace(); traceErr != nil {
						return traceErr
					}
					return checkScenarioAssertionsOnError(cmd, sc, err)
				}
				if traceErr := maybeTrace(); traceErr != nil {
					return traceErr
				}
				return err
			default:
				return fmt.Errorf("simulate: unexpected submit status %v", sub.Status)
			}
			continue

		case engine.RunFailed:
			err := res.Err
			if sc.ExpectErrorContains != "" {
				if traceErr := maybeTrace(); traceErr != nil {
					return traceErr
				}
				return checkScenarioAssertionsOnError(cmd, sc, err)
			}

			if traceErr := maybeTrace(); traceErr != nil {
				return traceErr
			}
			return err

		default:
			if err := maybeTrace(); err != nil {
				return err
			}
			return fmt.Errorf("simulate: unexpected run status %v", res.Status)
		}
	}

	if err := maybeTrace(); err != nil {
		return err
	}
	return fmt.Errorf("simulate: exceeded maxSteps=%d (possible loop)", stepLimit)
}

func printTrace(cmd *cobra.Command, in *engine.Instance, tf traceFormat) error {
	switch tf {
	case traceJSON:
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(in.Events())
	default:
		for _, ev := range in.Events() {
			if _, err := fmt.Fprintln(cmd.OutOrStdout(), ev.String()); err != nil {
				return err
			}
		}
		return nil
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
