package engine

import (
	"errors"
	"testing"

	"github.com/justinush/maestro/internal/definition"
)

func TestEngine_HappyPath_MinimalWorkflow(t *testing.T) {
	t.Parallel()

	def := &definition.WorkflowDefinition{
		SchemaVersion:   "0.1",
		ID:              "t",
		Version:         "1",
		InitialStepID:   "collect",
		TerminalStepIDs: []string{"done"},
		Steps: []definition.Step{
			{
				ID:              "collect",
				Kind:            definition.StepKindHuman,
				PresentationRef: "forms/x@v1",
				OnEnter: []definition.Action{
					{
						Type: "stub",
						ID:   "init",
						Params: mustRawJSON(t, map[string]any{
							"set": map[string]any{
								"stage": "collect",
							},
						}),
					},
				},
			},
			{
				ID:   "run",
				Kind: definition.StepKindAction,
				OnEnter: []definition.Action{
					{
						Type: "stub",
						ID:   "mark",
						Params: mustRawJSON(t, map[string]any{
							"set": map[string]any{
								"checksStarted": true,
							},
						}),
					},
				},
			},
			{
				ID:   "done",
				Kind: definition.StepKindEnd,
			},
		},
		Transitions: []definition.Transition{
			{From: "collect", To: "run", Priority: 0},
			{From: "run", To: "done", Priority: 0, When: `variables.checksStarted == true`},
		},
	}

	in, err := NewInstance(def, Options{})
	if err != nil {
		t.Fatalf("NewInstance: %v", err)
	}

	// First run should stop at human step needing input.
	if err := in.RunUntilBlocked(); !errors.Is(err, ErrNeedsInput) {
		t.Fatalf("RunUntilBlocked: want ErrNeedsInput, got %v", err)
	}
	if got := in.CurrentStepID(); got != "collect" {
		t.Fatalf("CurrentStepID: want %q, got %q", "collect", got)
	}

	// User submits input; engine advances to action step but does not auto-run until we call RunUntilBlocked again.
	if err := in.SubmitInput(map[string]any{"fullName": "Justin"}); err != nil {
		t.Fatalf("SubmitInput: %v", err)
	}
	if got := in.CurrentStepID(); got != "run" {
		t.Fatalf("CurrentStepID after SubmitInput: want %q, got %q", "run", got)
	}

	// Next run should execute action step and complete.
	if err := in.RunUntilBlocked(); !errors.Is(err, ErrWorkflowCompleted) {
		t.Fatalf("RunUntilBlocked: want ErrWorkflowCompleted, got %v", err)
	}
	if got := in.CurrentStepID(); got != "done" {
		t.Fatalf("CurrentStepID at end: want %q, got %q", "done", got)
	}

	vars := in.Variables()
	if vars["stage"] != "collect" {
		t.Fatalf("vars[stage]: want %q, got %#v", "collect", vars["stage"])
	}
	if vars["fullName"] != "Justin" {
		t.Fatalf("vars[fullName]: want %q, got %#v", "Justin", vars["fullName"])
	}
	if vars["checksStarted"] != true {
		t.Fatalf("vars[checksStarted]: want true, got %#v", vars["checksStarted"])
	}
}

func TestEngine_NoMatchingTransition(t *testing.T) {
	t.Parallel()

	def := &definition.WorkflowDefinition{
		SchemaVersion:   "0.1",
		ID:              "t",
		Version:         "1",
		InitialStepID:   "a",
		TerminalStepIDs: []string{"end"},
		Steps: []definition.Step{
			{ID: "a", Kind: definition.StepKindAction},
			{ID: "end", Kind: definition.StepKindEnd},
		},
		Transitions: []definition.Transition{
			{From: "a", To: "end", Priority: 0, When: "false"},
		},
	}

	in, err := NewInstance(def, Options{})
	if err != nil {
		t.Fatalf("NewInstance: %v", err)
	}

	err = in.RunUntilBlocked()
	if !errors.Is(err, ErrNoMatchingTransition) {
		t.Fatalf("RunUntilBlocked: want ErrNoMatchingTransition, got %v", err)
	}
}

func TestEngine_UnknownActionType(t *testing.T) {
	t.Parallel()

	def := &definition.WorkflowDefinition{
		SchemaVersion:   "0.1",
		ID:              "t",
		Version:         "1",
		InitialStepID:   "a",
		TerminalStepIDs: []string{"end"},
		Steps: []definition.Step{
			{
				ID:   "a",
				Kind: definition.StepKindAction,
				OnEnter: []definition.Action{
					{Type: "http", ID: "x"},
				},
			},
			{ID: "end", Kind: definition.StepKindEnd},
		},
		Transitions: []definition.Transition{
			{From: "a", To: "end", Priority: 0},
		},
	}

	in, err := NewInstance(def, Options{})
	if err != nil {
		t.Fatalf("NewInstance: %v", err)
	}

	err = in.RunUntilBlocked()
	if !errors.Is(err, ErrUnknownActionType) {
		t.Fatalf("RunUntilBlocked: want ErrUnknownActionType, got %v", err)
	}
}

func TestEngine_TransitionOrderingByPriorityThenDeclaration(t *testing.T) {
	t.Parallel()

	def := &definition.WorkflowDefinition{
		SchemaVersion:   "0.1",
		ID:              "t",
		Version:         "1",
		InitialStepID:   "a",
		TerminalStepIDs: []string{"end"},
		Steps: []definition.Step{
			{ID: "a", Kind: definition.StepKindAction},

			{ID: "b", Kind: definition.StepKindAction},

			{
				ID:   "c",
				Kind: definition.StepKindAction,
				OnEnter: []definition.Action{
					{Type: "http", ID: "should-fail-if-chosen"},
				},
			},

			{ID: "end", Kind: definition.StepKindEnd},
		},
		Transitions: []definition.Transition{
			// Same priority; declaration order must pick b before c.
			{From: "a", To: "b", Priority: 0, When: "true"},
			{From: "a", To: "c", Priority: 0, When: "true"},

			{From: "b", To: "end", Priority: 0},
			{From: "c", To: "end", Priority: 0},
		},
	}

	in, err := NewInstance(def, Options{})
	if err != nil {
		t.Fatalf("NewInstance: %v", err)
	}

	err = in.RunUntilBlocked()
	if errors.Is(err, ErrUnknownActionType) {
		t.Fatalf("RunUntilBlocked picked the wrong transition (a->c). Got %v", err)
	}
	if !errors.Is(err, ErrWorkflowCompleted) {
		t.Fatalf("RunUntilBlocked: want ErrWorkflowCompleted, got %v", err)
	}
	if got := in.CurrentStepID(); got != "end" {
		t.Fatalf("CurrentStepID: want %q, got %q", "end", got)
	}
}

func TestEngine_SubmitInput_EnforcesInputSchema(t *testing.T) {
	t.Parallel()

	def := &definition.WorkflowDefinition{
		SchemaVersion:   "0.1",
		ID:              "t",
		Version:         "1",
		InitialStepID:   "collect",
		TerminalStepIDs: []string{"done"},
		Steps: []definition.Step{
			{
				ID:              "collect",
				Kind:            definition.StepKindHuman,
				PresentationRef: "forms/x@v1",
				InputSchema: mustRawJSON(t, map[string]any{
					"type":     "object",
					"required": []any{"fullName"},
					"properties": map[string]any{
						"fullName": map[string]any{
							"type": "string",
						},
					},
					"additionalProperties": false,
				}),
			},
			{ID: "done", Kind: definition.StepKindEnd},
		},
		Transitions: []definition.Transition{
			{From: "collect", To: "done", Priority: 0},
		},
	}

	in, err := NewInstance(def, Options{})
	if err != nil {
		t.Fatalf("NewInstance: %v", err)
	}
	if err := in.RunUntilBlocked(); !errors.Is(err, ErrNeedsInput) {
		t.Fatalf("RunUntilBlocked: want ErrNeedsInput, got %v", err)
	}

	// Missing required fullName => should fail.
	err = in.SubmitInput(map[string]any{})
	if err == nil {
		t.Fatal("expected input schema error")
	}
	var ive *InputValidationError
	if !errors.As(err, &ive) {
		t.Fatalf("expected InputValidationError, got %T: %v", err, err)
	}
	if ive.StepID != "collect" {
		t.Fatalf("StepID: want %q, got %q", "collect", ive.StepID)
	}

	// Valid payload should pass.
	if err := in.SubmitInput(map[string]any{"fullName": "Justin"}); err != nil {
		t.Fatalf("SubmitInput(valid): %v", err)
	}
}

func mustRawJSON(t *testing.T, v any) definition.RawJSON {
	t.Helper()
	b, err := marshalToRawJSON(v)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
