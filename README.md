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

## Concepts (v0.1)

- **Definition**: YAML/JSON describing `steps` and `transitions`.
- **Step kinds**:
  - `human`: blocks until input is submitted
  - `action`: runs `onEnter`/`onExit` actions and auto-advances by transitions
  - `end`: terminal step (must be listed in `terminalStepIds`)
- **Guards**: `transitions[].when` is CEL. Today the runtime exposes `variables` to guards.
- **Variables**: a dynamic bag updated by actions and user input.
- **Input schema**: `steps[].inputSchema` is JSON Schema validated on `SubmitInput`.

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
make smoke         # validate + simulate (positive + negative scenarios)
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
