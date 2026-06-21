package validate

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestWorkflow(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		dataFile   string
		wantErrSub string
		opts       Options
	}{
		{
			name:       "valid_minimal",
			dataFile:   "valid_minimal.yaml",
			wantErrSub: "",
		},
		{
			name:       "schema_human_missing_presentationRef",
			dataFile:   "invalid_schema_human_no_presentation.yaml",
			wantErrSub: "schema validation",
		},
		{
			name:       "graph_unreachable_step",
			dataFile:   "invalid_graph_unreachable.yaml",
			wantErrSub: "not reachable",
		},
		{
			name:       "step_metadata_duplicate_label",
			dataFile:   "invalid_step_metadata_duplicate_label.yaml",
			wantErrSub: "duplicate label",
		},
		{
			name:       "graph_duplicate_transition",
			dataFile:   "invalid_graph_duplicate_transition.yaml",
			wantErrSub: "duplicate transition",
		},
		{
			name:       "cel_invalid_syntax",
			dataFile:   "invalid_cel_when.yaml",
			wantErrSub: "cel:",
		},
		{
			name:       "stub_unknown_params_field",
			dataFile:   "invalid_stub_params_unknown_key.yaml",
			wantErrSub: "params:",
		},
		{
			name:       "http_params_missing_url",
			dataFile:   "invalid_http_params_missing_url.yaml",
			wantErrSub: "url",
		},
		{
			name:       "inputschema_invalid_compile",
			dataFile:   "invalid_inputschema_compile.yaml",
			wantErrSub: "inputSchema",
		},
		{
			name:       "valid_custom_action_with_allowlist",
			dataFile:   "valid_custom_action.yaml",
			wantErrSub: "",
			opts: Options{
				AllowedActionTypes: []string{"vendor-create-session"},
			},
		},
		{
			name:       "invalid_custom_action_without_allowlist",
			dataFile:   "invalid_custom_action_type.yaml",
			wantErrSub: "schema validation",
		},
		{
			name:       "invalid_custom_action_unlisted_type",
			dataFile:   "valid_custom_action.yaml",
			wantErrSub: "schema validation",
			opts: Options{
				AllowedActionTypes: []string{"other-custom-action"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join("testdata", tt.dataFile)
			err := Workflow(path, tt.opts)
			if tt.wantErrSub == "" {
				if err != nil {
					t.Fatalf("Workflow: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected error")
			}
			if !strings.Contains(err.Error(), tt.wantErrSub) {
				t.Fatalf("error %q does not contain %q", err.Error(), tt.wantErrSub)
			}
		})
	}
}
