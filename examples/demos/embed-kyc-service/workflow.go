package main

import (
	_ "embed"

	"github.com/justinush/maestro/pkg/maestro"
)

//go:embed workflow.yaml
var workflowYAML []byte

// Embedded workflow: decode + validate (same checks as maestro.Load for a file).
func demoRuntime() (*maestro.Runtime, error) {
	return maestro.LoadYAML(workflowYAML)
}
