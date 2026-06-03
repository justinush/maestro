package workflow

import (
	"fmt"

	"github.com/justinush/maestro/pkg/definition"
	"github.com/justinush/maestro/pkg/run"
)

// checkRecordDefinitionMatch returns ErrDefinitionMismatch when rec and def disagree.
func checkRecordDefinitionMatch(rec *run.RunRecord, def *definition.WorkflowDefinition) error {
	if rec == nil {
		return fmt.Errorf("workflow: nil run record")
	}
	if def == nil {
		return fmt.Errorf("workflow: nil definition")
	}
	if rec.WorkflowID == def.ID && rec.WorkflowVersion == def.Version {
		return nil
	}
	return fmt.Errorf(
		"%w: run workflow id %q version %q, loaded definition id %q version %q",
		ErrDefinitionMismatch,
		rec.WorkflowID,
		rec.WorkflowVersion,
		def.ID,
		def.Version,
	)
}
