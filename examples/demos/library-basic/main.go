package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/justinush/maestro/pkg/definition"
	"github.com/justinush/maestro/pkg/engine"
	"github.com/justinush/maestro/pkg/validate"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <workflow.yaml>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "example: %s examples/workflows/workflow-v0-minimal.yaml\n", os.Args[0])
		os.Exit(2)
	}
	path := os.Args[1]

	// 1. Decode the workflow definition from YAML/JSON.
	def, err := definition.DecodeFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "decode: %v\n", err)
		os.Exit(1)
	}

	// 2. Validate before creating a runtime instance (same checks as maestro validate).
	if err := validate.WorkflowDefinition(def, validate.Options{}); err != nil {
		fmt.Fprintf(os.Stderr, "validate: %v\n", err)
		os.Exit(1)
	}

	// 3. Create an in-memory workflow instance.
	in, err := engine.NewInstance(def, engine.Options{
		ActionRegistry: engine.DefaultRegistry(),
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "instance: %v\n", err)
		os.Exit(1)
	}

	// 4. Drive until the workflow completes or pauses for external input (e.g. human step → ErrNeedsInput).
	for {
		err := in.RunUntilBlocked()
		switch {
		case errors.Is(err, engine.ErrWorkflowCompleted):
			fmt.Printf("completed at step %q\n", in.CurrentStepID())
			return
		case errors.Is(err, engine.ErrNeedsInput):
			fmt.Printf("blocked at %q (needs input — see examples/demos/embed-kyc-service)\n", in.CurrentStepID())
			return
		case err != nil:
			fmt.Fprintf(os.Stderr, "run: %v\n", err)
			os.Exit(1)
		}
	}
}
