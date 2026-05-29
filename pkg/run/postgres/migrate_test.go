package postgres_test

import (
	"strings"
	"testing"

	"github.com/justinush/maestro/pkg/run/postgres"
)

func TestSchemaDDL(t *testing.T) {
	ddl := postgres.SchemaDDL()
	if !strings.Contains(ddl, "workflow_runs") {
		t.Fatalf("SchemaDDL missing workflow_runs: %q", ddl)
	}
	if !strings.Contains(ddl, "run_id") {
		t.Fatalf("SchemaDDL missing run_id: %q", ddl)
	}
}
