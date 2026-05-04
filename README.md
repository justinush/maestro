# Maestro

Maestro is a small workflow platform: define graphs in YAML or JSON, validate them strictly, and run instances in memory as execution lands.

## Requirements

- [Go](https://go.dev/dl/) **1.26+** (see `go.mod`)

## Quick start

```bash
go build -o maestro ./cmd/maestro
./maestro validate -f examples/workflow-v0-minimal.yaml
```

Without a local binary:

```bash
go run ./cmd/maestro validate -f examples/workflow-v0-minimal.yaml
```

Optional schema override and richer errors:

```bash
./maestro validate -f path/to/workflow.yaml --schema path/to/schema.json --verbose
```

Smoke the bundled example via Make:

```bash
make validate-example
```

## Development

```bash
make test          # or: go test ./...
make check         # lint + vet + test
make build         # binary under dist/
```

See `make help` for all targets.

## Roadmap / tech debt

Worth tracking as the codebase grows:

- **`simulate`**: command exists; full scenario-driven simulation is not wired yet.
- **Validate API**: `validate.Workflow` is file-based; a **definition-only** validate path would help callers that already hold a `WorkflowDefinition`.
- **CEL**: validate and the engine must stay aligned on **bindings** (`variables`, and future names).
- **JSON Schema**: custom root **`$id`** is supported; **external `$ref`** chains are not pre-loaded.
- **Engine tests**: expand coverage as advancement, variables, and actions land.
- **Logging**: structured logging belongs with a long-running service; the CLI stays stderr-oriented for now.
