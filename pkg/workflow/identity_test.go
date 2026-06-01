package workflow

import (
	"errors"
	"strings"
	"testing"

	"github.com/justinush/maestro/pkg/definition"
	"github.com/justinush/maestro/pkg/run"
)

func TestCheckRecordDefinitionMatch_ok(t *testing.T) {
	t.Parallel()
	rec := &run.RunRecord{WorkflowID: "a", WorkflowVersion: "1"}
	def := &definition.WorkflowDefinition{ID: "a", Version: "1"}
	if err := checkRecordDefinitionMatch(rec, def); err != nil {
		t.Fatal(err)
	}
}

func TestCheckRecordDefinitionMatch_mismatch(t *testing.T) {
	t.Parallel()
	rec := &run.RunRecord{WorkflowID: "a", WorkflowVersion: "1"}
	def := &definition.WorkflowDefinition{ID: "a", Version: "2"}
	err := checkRecordDefinitionMatch(rec, def)
	if !errors.Is(err, ErrDefinitionMismatch) {
		t.Fatalf("want ErrDefinitionMismatch, got %v", err)
	}
	msg := err.Error()
	for _, sub := range []string{
		`run workflow id "a"`,
		`version "1"`,
		`loaded definition id "a"`,
		`version "2"`,
	} {
		if !strings.Contains(msg, sub) {
			t.Fatalf("error %q missing %q", msg, sub)
		}
	}
}
