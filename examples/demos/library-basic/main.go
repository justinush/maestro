package main

import (
	"fmt"
	"os"

	"github.com/justinush/maestro/pkg/engine"
	"github.com/justinush/maestro/pkg/maestro"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <workflow.yaml>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "example: %s examples/workflows/workflow-v0-minimal.yaml\n", os.Args[0])
		os.Exit(2)
	}

	// 1. Decode + validate (same checks as maestro validate).
	rt, err := maestro.Load(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	if def := rt.Definition(); def != nil {
		fmt.Printf("loaded workflow %q (version %q)\n", def.ID, def.Version)
	}

	// 2. Create an in-memory workflow instance (default stub registry when ActionRegistry is unset).
	in, err := rt.NewInstance(maestro.InstanceOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "instance: %v\n", err)
		os.Exit(1)
	}

	// 3. One drive: engine runs internally until blocked, completed, or failed.
	res := in.RunUntilBlocked()
	switch res.Status {
	case engine.RunCompleted:
		fmt.Printf("completed at step %q\n", res.StepID)
		return
	case engine.RunBlocked:
		fmt.Printf("blocked at %q (needs input — see examples/demos/embed-kyc-service)\n", res.StepID)
		return
	case engine.RunFailed:
		fmt.Fprintf(os.Stderr, "run: %v\n", res.Err)
		os.Exit(1)
	default:
		fmt.Fprintf(os.Stderr, "run: unexpected status %v\n", res.Status)
		os.Exit(1)
	}
}
