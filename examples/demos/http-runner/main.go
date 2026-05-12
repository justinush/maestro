package main

import (
	"bytes"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/justinush/maestro/pkg/definition"
	"github.com/justinush/maestro/pkg/engine"
	"github.com/justinush/maestro/pkg/validate"
	"gopkg.in/yaml.v3"
)

//go:embed workflow.yaml
var workflowTemplate []byte

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	fmt.Println("ok: workflow reached approved via HTTP runner + guard")
}

func run() error {
	srv := newVendorMock()
	defer srv.Close()

	client := srv.Client()
	if client == nil {
		return fmt.Errorf("httptest server client is nil")
	}

	yamlBody := strings.Replace(string(workflowTemplate), "__HTTP_BASE__", srv.URL, 1)

	dec := yaml.NewDecoder(bytes.NewReader([]byte(yamlBody)))
	dec.KnownFields(true)
	var def definition.WorkflowDefinition
	if err := dec.Decode(&def); err != nil {
		return fmt.Errorf("parse workflow: %w", err)
	}
	d := &def

	if err := validate.WorkflowDefinition(d, validate.Options{}); err != nil {
		return fmt.Errorf("validate: %w", err)
	}

	reg := engine.RegistryWithHTTP(client)

	in, err := engine.NewInstance(d, engine.Options{
		RunID:          "http_demo",
		ActionRegistry: reg,
	})
	if err != nil {
		return fmt.Errorf("new instance: %w", err)
	}

	if err := in.RunUntilBlocked(); err != nil {
		if !errors.Is(err, engine.ErrWorkflowCompleted) {
			return fmt.Errorf("run: %w", err)
		}
	}

	if in.CurrentStepID() != "approved" {
		return fmt.Errorf("want terminal step approved, got %q", in.CurrentStepID())
	}

	v, ok := in.Variables()["vendor"]
	if !ok {
		return fmt.Errorf("missing vendor resultVariable")
	}
	fmt.Printf("vendor snapshot: %#v\n", v)

	for _, ev := range in.Events() {
		fmt.Println(ev.String())
	}
	return nil
}
