package workflow_test

import (
	"fmt"
	"log"

	"github.com/justinush/maestro/pkg/maestro"
	"github.com/justinush/maestro/pkg/workflow"
)

const minimalYAML = `
schemaVersion: "0.1"
id: demo
version: "1"
initialStepId: start
terminalStepIds: [done]
steps:
  - { id: start, kind: action }
  - { id: done, kind: end }
transitions:
  - { from: start, to: done, priority: 0 }
`

func ExampleNewRegistry() {
	reg := workflow.NewRegistry()
	rt, err := maestro.LoadYAML([]byte(minimalYAML))
	if err != nil {
		log.Fatal(err)
	}
	if err := reg.Register(rt); err != nil {
		log.Fatal(err)
	}
	fmt.Println(reg.Contains(workflow.Key{ID: "demo", Version: "1"}))
	// Output: true
}

func ExampleRegistry_NewInstance() {
	reg := workflow.NewRegistry()
	rt, err := maestro.LoadYAML([]byte(minimalYAML))
	if err != nil {
		log.Fatal(err)
	}
	if err := reg.Register(rt); err != nil {
		log.Fatal(err)
	}
	_, err = reg.NewInstance(workflow.Key{ID: "demo", Version: "1"}, maestro.InstanceOptions{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(true)
	// Output: true
}
