package main

import (
	_ "embed"
	"strings"

	"github.com/justinush/maestro/pkg/maestro"
)

//go:embed workflow.yaml
var workflowYAML []byte

// demoRuntime substitutes the mock vendor base URL, then decode + validate (same as maestro.Load).
func demoRuntime(vendorBaseURL string) (*maestro.Runtime, error) {
	yamlBody := strings.Replace(string(workflowYAML), "__HTTP_BASE__", vendorBaseURL, 1)
	return maestro.LoadYAML([]byte(yamlBody))
}
