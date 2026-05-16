package main

import (
	"context"
	"fmt"

	"github.com/justinush/maestro/pkg/engine"
	"github.com/justinush/maestro/pkg/maestro"
	"github.com/justinush/maestro/pkg/run"
)

const demoRunID = "run_demo"

func runDemo() error {
	ctx := context.Background()

	// 1. Load workflow (embedded YAML - see workflow.go; same idea as maestro.Load(path)).
	rt, err := demoRuntime()
	if err != nil {
		return err
	}
	def := rt.Definition()

	store := run.NewMemoryStore()
	fmt.Printf("created run: %s\n", demoRunID)

	// 2. Start a new run (default stub registry when ActionRegistry is unset).
	in, err := rt.NewInstance(maestro.InstanceOptions{RunID: demoRunID})
	if err != nil {
		return fmt.Errorf("new instance: %w", err)
	}

	// 3. Drive until the engine pauses on a human step.
	res := in.RunUntilBlocked()
	if res.Status != engine.RunBlocked {
		if res.Err != nil {
			return fmt.Errorf("first run: want RunBlocked, got %v: %w", res.Status, res.Err)
		}
		return fmt.Errorf("first run: want RunBlocked, got %v", res.Status)
	}
	fmt.Printf("blocked at: %s\n", res.StepID)

	// 4. First request ends: persist snapshot while blocked (your DB / store).
	if err := persistNewRun(ctx, store, in, def); err != nil {
		return err
	}

	// 5. Second request: reload from the store, like a later API handler would.
	in, err = restoreRun(ctx, rt, store, demoRunID)
	if err != nil {
		return err
	}
	fmt.Println("restored run from store")

	// 6. Human input on the blocked step.
	fmt.Println()
	fmt.Println("submitting profile input...")
	sub := in.SubmitInput(map[string]any{"fullName": "Demo User"})
	switch sub.Status {
	case engine.SubmitAdvanced:
	case engine.SubmitFailed:
		return fmt.Errorf("submit input: %w", sub.Err)
	default:
		return fmt.Errorf("submit input: expected SubmitAdvanced, got %v", sub.Status)
	}
	fmt.Printf("continued to: %s\n", sub.StepID)

	// 7. Drive to terminal step, then save final state (one call is enough for this graph; loop for longer workflows).
	for {
		res := in.RunUntilBlocked()
		switch res.Status {
		case engine.RunBlocked:
			return fmt.Errorf("unexpected block at step %q", res.StepID)
		case engine.RunCompleted:
			fmt.Printf("completed at: %s\n", res.StepID)
			if err := saveRun(ctx, store, demoRunID, in, def); err != nil {
				return err
			}
			fmt.Println("saved final run state")
			fmt.Println()
			fmt.Println("trace:")
			for _, ev := range in.Events() {
				fmt.Println(ev.String())
			}
			return nil
		case engine.RunFailed:
			return fmt.Errorf("run: %w", res.Err)
		default:
			return fmt.Errorf("run: unexpected status %v", res.Status)
		}
	}
}
