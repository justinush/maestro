# Maestro

A scalable workflow engine for KYC automation.

## Requirements

- [Go](https://go.dev/dl/) **1.26+** (see `go.mod`)

## Quick start

```bash
go build -o maestro ./cmd/maestro
./maestro validate -f examples/workflow-v0-minimal.yaml
```

Run without building a binary:

```bash
go run ./cmd/maestro validate -f examples/workflow-v0-minimal.yaml
```

Simulate a run from a scenario file (supports assertions):

```bash
go run ./cmd/maestro simulate -s examples/scenario-minimal.yaml
```

Print an execution trace (text or JSON):

```bash
go run ./cmd/maestro simulate -s examples/scenario-minimal.yaml --trace
go run ./cmd/maestro simulate -s examples/scenario-minimal.yaml --trace --trace-format json
go run ./cmd/maestro simulate -s examples/scenario-minimal.yaml --trace-guards
```

Optional schema override and richer errors:

```bash
./maestro validate -f path/to/workflow.yaml --schema path/to/schema.json --verbose
```

Smoke the bundled example via Make:

```bash
make validate-example
```

## Library

Stable imports for embedding the engine in your own service or tool:

- **`github.com/justinush/maestro/pkg/definition`** — workflow types and **`DecodeFile`** (strict YAML/JSON).
- **`github.com/justinush/maestro/pkg/engine`** — **`NewInstance`**, **`RunUntilBlocked`**, **`SubmitInput`**, **`Registry`**, **`ActionRunner`**, trace **`Event`** values, and sentinel errors.
- **`github.com/justinush/maestro/pkg/validate`** — **`WorkflowFile`** and **`WorkflowDefinition`** (same checks as **`maestro validate`**).

Typical flow: **`definition.DecodeFile`** → optional **`validate.WorkflowDefinition`** → **`engine.NewInstance`** → **`RunUntilBlocked`** / **`SubmitInput`**.

Minimal program:

```bash
go run ./examples/library examples/workflow-v0-minimal.yaml
```

After you publish tags, pin with **`require github.com/justinush/maestro v0.x.y`** in the consumer module.

## Concepts (v0.1)

- **Definition**: YAML/JSON describing `steps` and `transitions`.

- **Step kinds**:
  - `human`: blocks until input is submitted
  - `action`: runs `onEnter`/`onExit` actions and auto-advances by transitions
  - `end`: terminal step (must be listed in `terminalStepIds`)

### Exit guards

Leaving a step means choosing a **transition** out of that step. Optional CEL on **`transitions[].when`** is an **exit guard**: it must evaluate to true for that edge to be followed. Omitting or leaving `when` empty means that transition is **unconditional** (subject to engine ordering: `priority`, then declaration order among ties).

Authoring rule of thumb: **`onEnter` / `onExit` actions** and **`SubmitInput`** should populate **`variables`** so exit guards read a small, stable shape—avoid hiding state only inside presentation layers.

The engine currently exposes **`variables`** to CEL; keep guards aligned with whatever your actions and inputs write there.

### Variables

**`variables`** is the instance state bag: top-level keys are set by **`stub`**, **`http`** (via `resultVariable`), and a **shallow merge** on **`SubmitInput`**. Nested maps are fine as **values**; updating nested fields is still “replace the whole top-level key” unless you model finer paths yourself.

**Naming convention (recommended, not enforced by the engine):**

| Area | Suggested top-level key | Contents |
|------|---------------------------|----------|
| Partner / corridor sync | `partner` | e.g. `status`, error codes |
| Customer relationship / base capture | `profile` or flat keys you prefer | facts agreed at COR-style steps |
| Liveness / face | `liveness` | e.g. `passed`, vendor hints |
| Address / POA | `address` | e.g. `status`, `reviewRequired` |
| Integrations / probes | `integration`, `integrationHttpTodo`, … | HTTP echo objects, staging flags |

Re-use the same keys across **definitions**, **`simulate`** `initialVariables`, and **`expectVariables`** so validation (CEL compile) and scenarios stay easy to reason about.

### Input schema

**`steps[].inputSchema`** is JSON Schema, validated on **`SubmitInput`** for `human` steps.

## CLI

### `maestro validate`

Validates a workflow definition by:

- decoding YAML/JSON strictly
- JSON Schema validation (embedded v0.1 schema, or `--schema`)
- graph checks (ids, terminals, reachability, transitions)
- CEL compile checks for `when`
- stub params shape checks
- `inputSchema` compilation checks

### `maestro simulate`

Runs the engine using a scenario file (`.yaml`/`.json`) that can:

- set initial variables
- provide scripted inputs for human steps
- assert completion and variables (positive cases)
- assert expected error substring (negative cases)
- print an execution trace (`--trace`, `--trace-format json`, `--trace-guards`)

## Development

```bash
make test          # or: go test ./...
make check         # lint + vet + test
make build         # binary under dist/
make smoke         # validate + simulate (minimal + negative + SG portrait)
```

See `make help` for all targets.

## Roadmap / tech debt

Worth tracking as the codebase grows:

- **Stop reasons**: the engine currently uses sentinel errors (`ErrNeedsInput`, `ErrWorkflowCompleted`) as control flow; consider `(Status, error)` when integrations grow.
- **CEL performance**: guards are cached per transition; consider caching a shared env and adding cost limits if needed for long-running services.
- **CEL**: validate and the engine must stay aligned on **bindings** (`variables`, and future names).
- **JSON Schema**: custom root **`$id`** is supported; **external `$ref`** chains are not pre-loaded.
- **Scenario assertions**: keep expanding example scenarios as contracts (positive and negative cases).
- **Engine trace**: trace is implemented; consider timestamps, persistence, and stable JSON output contracts.
- **Logging**: structured logging belongs with a long-running service; the CLI stays stderr-oriented for now.
