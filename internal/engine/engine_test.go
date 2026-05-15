package engine

import (
	"errors"
	"net/http"
	"net/http/httptest"
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
	res := in.RunUntilBlocked()
	if res.Status != RunBlocked {
		t.Fatalf("RunUntilBlocked: want RunBlocked, got %v err=%v", res.Status, res.Err)
	}
	if got := res.StepID; got != "collect" {
		t.Fatalf("StepID: want %q, got %q", "collect", got)
	}
	if got := in.CurrentStepID(); got != "collect" {
		t.Fatalf("CurrentStepID: want %q, got %q", "collect", got)
	}

	// User submits input; engine advances to action step but does not auto-run until we call RunUntilBlocked again.
	sub := in.SubmitInput(map[string]any{"fullName": "Justin"})
	if sub.Status != SubmitAdvanced {
		t.Fatalf("SubmitInput: want SubmitAdvanced, got %v err=%v", sub.Status, sub.Err)
	}
	if got := sub.StepID; got != "run" {
		t.Fatalf("StepID after SubmitInput: want %q, got %q", "run", got)
	}

	// Next run should execute action step and complete.
	res = in.RunUntilBlocked()
	if res.Status != RunCompleted {
		t.Fatalf("RunUntilBlocked: want RunCompleted, got %v err=%v", res.Status, res.Err)
	}
	if got := res.StepID; got != "done" {
		t.Fatalf("StepID at end: want %q, got %q", "done", got)
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

	res := in.RunUntilBlocked()
	if res.Status != RunFailed {
		t.Fatalf("RunUntilBlocked: want RunFailed, got %v", res.Status)
	}
	if !errors.Is(res.Err, ErrNoMatchingTransition) {
		t.Fatalf("RunUntilBlocked: want ErrNoMatchingTransition, got %v", res.Err)
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

	res := in.RunUntilBlocked()
	if res.Status != RunFailed {
		t.Fatalf("RunUntilBlocked: want RunFailed, got %v", res.Status)
	}
	if !errors.Is(res.Err, ErrUnknownActionType) {
		t.Fatalf("RunUntilBlocked: want ErrUnknownActionType, got %v", res.Err)
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

	res := in.RunUntilBlocked()
	if errors.Is(res.Err, ErrUnknownActionType) {
		t.Fatalf("RunUntilBlocked picked the wrong transition (a->c). Got %v", res.Err)
	}
	if res.Status != RunCompleted {
		t.Fatalf("RunUntilBlocked: want RunCompleted, got %v err=%v", res.Status, res.Err)
	}
	if got := res.StepID; got != "end" {
		t.Fatalf("StepID: want %q, got %q", "end", got)
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
	res := in.RunUntilBlocked()
	if res.Status != RunBlocked {
		t.Fatalf("RunUntilBlocked: want RunBlocked, got %v err=%v", res.Status, res.Err)
	}

	// Missing required fullName => should fail.
	sub := in.SubmitInput(map[string]any{})
	if sub.Status != SubmitFailed {
		t.Fatalf("SubmitInput: want SubmitFailed, got %v", sub.Status)
	}
	var ive *InputValidationError
	if !errors.As(sub.Err, &ive) {
		t.Fatalf("expected InputValidationError, got %T: %v", sub.Err, sub.Err)
	}
	if ive.StepID != "collect" {
		t.Fatalf("StepID: want %q, got %q", "collect", ive.StepID)
	}

	// Valid payload should pass and advance.
	sub = in.SubmitInput(map[string]any{"fullName": "Justin"})
	if sub.Status != SubmitAdvanced {
		t.Fatalf("SubmitInput(valid): want SubmitAdvanced, got %v err=%v", sub.Status, sub.Err)
	}
}

func TestEngine_SubmitInput_NoMatchingTransition_StaysOnHumanStep(t *testing.T) {
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
			{
				From:     "collect",
				To:       "done",
				Priority: 0,
				When:     "false",
			},
		},
	}

	in, err := NewInstance(def, Options{})
	if err != nil {
		t.Fatalf("NewInstance: %v", err)
	}

	res := in.RunUntilBlocked()
	if res.Status != RunBlocked {
		t.Fatalf("RunUntilBlocked: want RunBlocked, got %v err=%v", res.Status, res.Err)
	}
	if got := in.CurrentStepID(); got != "collect" {
		t.Fatalf("CurrentStepID: want %q, got %q", "collect", got)
	}

	sub := in.SubmitInput(map[string]any{"fullName": "Justin"})
	if sub.Status != SubmitStayOnStep {
		t.Fatalf("SubmitInput: want SubmitStayOnStep, got %v err=%v", sub.Status, sub.Err)
	}
	if got := sub.StepID; got != "collect" {
		t.Fatalf("StepID after SubmitInput: want %q, got %q", "collect", got)
	}
}

type testSetVarRunner struct {
	key string
	val string
}

func (r testSetVarRunner) Run(ctx ActionContext) error {
	ctx.Variables[r.key] = r.val
	return nil
}

func TestEngine_CustomActionRegistry(t *testing.T) {
	t.Parallel()

	reg := NewRegistry()
	reg.MustRegister("stub", NewStubRunner())
	reg.MustRegister("setVar", testSetVarRunner{key: "custom", val: "ok"})

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
					{Type: "setVar", ID: "x"},
				},
			},
			{ID: "end", Kind: definition.StepKindEnd},
		},
		Transitions: []definition.Transition{
			{From: "a", To: "end", Priority: 0},
		},
	}

	in, err := NewInstance(def, Options{ActionRegistry: reg})
	if err != nil {
		t.Fatalf("NewInstance: %v", err)
	}
	res := in.RunUntilBlocked()
	if res.Status != RunCompleted {
		t.Fatalf("RunUntilBlocked: want RunCompleted, got %v err=%v", res.Status, res.Err)
	}
	if in.Variables()["custom"] != "ok" {
		t.Fatalf("expected custom variable set by runner")
	}
}

func TestEngine_HTTPRunner(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Echo-Method", r.Method)
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)

	reg := NewRegistry()
	reg.MustRegister("stub", NewStubRunner())
	reg.MustRegister("http", NewHTTPRunner(srv.Client()))

	def := &definition.WorkflowDefinition{
		SchemaVersion:   "0.1",
		ID:              "t",
		Version:         "1",
		InitialStepID:   "call",
		TerminalStepIDs: []string{"end"},
		Steps: []definition.Step{
			{
				ID:   "call",
				Kind: definition.StepKindAction,
				OnEnter: []definition.Action{
					{
						Type: "http",
						ID:   "fetch",
						Params: mustRawJSON(t, map[string]any{
							"url":            srv.URL + "/path",
							"resultVariable": "httpResult",
							"timeoutSeconds": 5,
						}),
					},
				},
			},
			{ID: "end", Kind: definition.StepKindEnd},
		},
		Transitions: []definition.Transition{
			{From: "call", To: "end", Priority: 0},
		},
	}

	in, err := NewInstance(def, Options{ActionRegistry: reg})
	if err != nil {
		t.Fatalf("NewInstance: %v", err)
	}
	res := in.RunUntilBlocked()
	if res.Status != RunCompleted {
		t.Fatalf("RunUntilBlocked: want RunCompleted, got %v err=%v", res.Status, res.Err)
	}

	v := in.Variables()["httpResult"]
	m, ok := v.(map[string]any)
	if !ok {
		t.Fatalf("httpResult type %T", v)
	}
	gotCode, ok := m["statusCode"].(int)
	if !ok || gotCode != http.StatusTeapot {
		t.Fatalf("statusCode: want %d, got %#v (%T)", http.StatusTeapot, m["statusCode"], m["statusCode"])
	}
	body, _ := m["body"].(string)
	if body != `{"ok":true}` {
		t.Fatalf("body: got %q", body)
	}
}

func TestEngine_SnapshotRoundTrip(t *testing.T) {
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
			{From: "a", To: "end", Priority: 0},
		},
	}

	in, err := NewInstance(def, Options{RunID: "run-1", InitialVariables: map[string]any{"k": 1}})
	if err != nil {
		t.Fatalf("NewInstance: %v", err)
	}
	res := in.RunUntilBlocked()
	if res.Status != RunCompleted {
		t.Fatalf("RunUntilBlocked: %v (status=%v)", res.Err, res.Status)
	}

	snap := in.Snapshot()
	if snap.CurrentStepID != "end" {
		t.Fatalf("snap step: %q", snap.CurrentStepID)
	}
	if snap.NextSeq < 1 {
		t.Fatalf("snap nextSeq: %d", snap.NextSeq)
	}

	in2, err := NewInstanceFromSnapshot(def, snap, Options{ActionRegistry: DefaultRegistry()})
	if err != nil {
		t.Fatalf("NewInstanceFromSnapshot: %v", err)
	}
	if in2.CurrentStepID() != snap.CurrentStepID {
		t.Fatalf("restored step: %q", in2.CurrentStepID())
	}
	if in2.RunID() != "run-1" {
		t.Fatalf("restored runId: %q", in2.RunID())
	}
	if got := in2.Variables()["k"]; got != float64(1) && got != 1 {
		// JSON round-trips numbers as float64; in-memory int stays int
		t.Fatalf("variables k: %#v", got)
	}
	if len(in2.Events()) != len(in.Events()) {
		t.Fatalf("events len: %d vs %d", len(in2.Events()), len(in.Events()))
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
