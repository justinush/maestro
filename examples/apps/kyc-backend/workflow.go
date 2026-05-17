package main

import (
	_ "embed"

	"github.com/justinush/maestro/pkg/maestro"
)

//go:embed workflow.yaml
var workflowYAML []byte

func loadRuntime() (*maestro.Runtime, error) {
	return maestro.LoadYAML(workflowYAML)
}
