package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/justinush/maestro/pkg/engine"
	"github.com/justinush/maestro/pkg/maestro"
)

const demoRunID = "http_demo"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	srv := newVendorMock()
	defer srv.Close()
	fmt.Println("mock vendor started")

	client := srv.Client()
	if client == nil {
		return fmt.Errorf("httptest server client is nil")
	}

	// Register the HTTP runner with the mock vendor client.
	// In a real service, this client would call your vendor or internal verification API.
	reg := engine.RegistryWithHTTP(client)

	// 1. Load workflow (embedded YAML with vendor base URL - see workflow.go).
	rt, err := demoRuntime(srv.URL)
	if err != nil {
		return err
	}
	if def := rt.Definition(); def != nil {
		fmt.Printf("loaded workflow %q (version %q)\n", def.ID, def.Version)
	}

	fmt.Println("running workflow with HTTP runner...")

	// 2. Start run with an action registry that includes the HTTP runner.
	in, err := rt.NewInstance(maestro.InstanceOptions{
		RunID:          demoRunID,
		ActionRegistry: reg,
	})
	if err != nil {
		return fmt.Errorf("new instance: %w", err)
	}

	// 3. Drive to completion (action step runs HTTP onEnter, guard checks statusCode).
	res := in.RunUntilBlocked()
	switch res.Status {
	case engine.RunCompleted:
		// ok
	case engine.RunFailed:
		return fmt.Errorf("run: %w", res.Err)
	default:
		return fmt.Errorf("run: unexpected status %v", res.Status)
	}

	if res.StepID != "approved" {
		return fmt.Errorf("want terminal step approved, got %q", res.StepID)
	}
	fmt.Printf("completed at: %s\n", res.StepID)

	if err := printVendorResult(in); err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("trace:")
	for _, ev := range in.Events() {
		fmt.Println(ev.String())
	}
	return nil
}

func printVendorResult(in *engine.Instance) error {
	raw, ok := in.Variables()["vendor"]
	if !ok {
		return fmt.Errorf("missing variables.vendor (HTTP resultVariable)")
	}
	vm, ok := raw.(map[string]any)
	if !ok {
		return fmt.Errorf("variables.vendor: expected map, got %T", raw)
	}

	fmt.Printf("vendor HTTP status: %v\n", vm["statusCode"])

	body, _ := vm["body"].(string)
	if body == "" {
		return nil
	}

	var payload struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err == nil && payload.Status != "" {
		fmt.Printf("vendor liveness status: %s\n", payload.Status)
		return nil
	}

	fmt.Printf("vendor response body: %s\n", body)
	return nil
}
