package maestro_test

import (
	"fmt"
	"log"

	"github.com/justinush/maestro/pkg/engine"
	"github.com/justinush/maestro/pkg/maestro"
)

const exampleWorkflowYAML = `
schemaVersion: "0.1"
id: example
version: "1"
initialStepId: start
terminalStepIds: [done]
steps:
  - id: start
    kind: action
  - id: done
    kind: end
transitions:
  - from: start
    to: done
    priority: 0
`

func ExampleLoadYAML() {
	rt, err := maestro.LoadYAML([]byte(exampleWorkflowYAML))
	if err != nil {
		log.Fatal(err)
	}
	in, err := rt.NewInstance(maestro.InstanceOptions{})
	if err != nil {
		log.Fatal(err)
	}
	res := in.RunUntilBlocked()
	fmt.Println(res.Status == engine.RunCompleted)
	// Output: true
}
