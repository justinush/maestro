# Maestro

Extensible workflow runtime for KYC and onboarding systems.

## The problem

Modern KYC and onboarding products tend to accumulate **workflow logic in application code**: branching rules, vendor outcomes, and market-specific paths spread across services and releases. That leads to:

- **Tight coupling** between “what the journey is” and “how we execute it today,” making changes risky and slow.
- **Hardcoded transitions** that are hard to review, diff, or reuse across corridors and entities.
- **Vendor-specific branching** mixed with core orchestration, so swapping or A/B testing integrations is painful.
- **Limited visibility** into how a case moved through the journey when debugging or auditing.
- **Harder customization** when each region or partner needs a slightly different path without a single source of truth.

As flows grow, the cost of change and the risk of regressions rise faster than the size of the team.

## The solution

Maestro separates **what the workflow is** (a declarative definition) from **how your stack runs it** (an embeddable runtime in Go).

It gives you:

- **Declarative definitions** (YAML/JSON): steps, transitions, and exit guards instead of ad hoc `if` chains for orchestration.
- **An embeddable engine** you can run from services, workers, or tests with the same behavior as the CLI.
- **Pluggable action runners** (for example `stub` and `http`) plus a registry for your own integrations.
- **Validation and simulation** so definitions are checked before production and scenarios can lock in expected paths.
- **Execution tracing** for observability while a run is in memory.
- **Snapshot-oriented persistence hooks** (`pkg/run` store + in-memory implementation) so you can persist and resume execution state without baking a database into the core.

The goal is not a no-code platform: it is a **small, explicit runtime** that teams can adopt incrementally and extend where it matters.

## Core concepts (v0.1)

- **Workflow definition**: YAML/JSON describing `steps` and `transitions` (revision **0.1** today).

- **Step kinds**:
  - `human`: blocks until input is submitted.
  - `action`: runs `onEnter` / `onExit` actions and auto-advances according to transitions.
  - `end`: terminal step (must be listed in `terminalStepIds`).

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

## Architecture

Maestro is a **library-first** pipeline: definitions and tooling sit above a small runtime that delegates side effects to **action runners**. Your product code owns persistence, identity, channels, and long-lived storage.

```
  Workflow definition (YAML / JSON)
              │
              ▼
       ┌──────────────┐
       │  Validation  │  JSON Schema, graph, CEL compile,
       │  + simulate  │  stub/http params, inputSchema
       └──────────────┘
              │
              ▼
       ┌──────────────┐
       │   Instance   │  current step, variables, exit guards,
       │   (engine)   │  RunUntilBlocked / SubmitInput
       └──────────────┘
              │
              ▼
       ┌──────────────┐
       │Action runners│  Registry: stub, http, your types
       └──────────────┘
              │
              ▼
  Your application (HTTP APIs, queues, pkg/run Store, UI)
```

The **`maestro`** CLI is a thin wrapper around the same **`pkg/validate`** and **`pkg/engine`** packages your service would import.

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

Full CI-style checks (minimal + negative scenarios + Singapore portrait):

```bash
make smoke
```

## Examples

### Minimal onboarding (`examples/workflow-v0-minimal.yaml`)

A three-step graph: **`collect-profile`** (human) → **`run-checks`** (action, stub sets `checksStarted`) → **`done`** (end). Transitions use CEL on the edge from **`run-checks`** to **`done`**.

Run it with **`examples/scenario-minimal.yaml`** (scripted input + assertions).

### Singapore portrait journey (`examples/kyc/sg/portrait/`)

A longer example: partner gate, customer relationship capture, stubbed integration snapshot, liveness, proof-of-address, and multiple **end** terminals (approved, partner rejected, POA refused, manual review). Bundled scenarios: **`scenario-happy.yaml`**, **`scenario-partner-rejected.yaml`**, **`scenario-poa-review.yaml`**. The default workflow uses **stubs only** for integration (no outbound HTTP in **`make smoke`**).

### Execution trace (text)

```bash
go run ./cmd/maestro simulate -s examples/scenario-minimal.yaml --trace
```

Example lines (sequence and event types; your step ids may differ):

```
0001 step.entered step="collect-profile"
0002 action.ran step="collect-profile" list=onEnter action="init-vars" type="stub"
0003 run.blocked step="collect-profile"
0004 input.accepted step="collect-profile" keys=["fullName"]
0005 transition.taken idx=0 from="collect-profile" to="run-checks"
0006 step.entered step="run-checks"
0007 action.ran step="run-checks" list=onEnter action="mark-checks-started" type="stub"
0008 transition.taken idx=1 from="run-checks" to="done"
0009 step.entered step="done"
0010 run.completed step="done"
ok (completed "done")
```

Use **`--trace-format json`** for machine-readable output, or **`--trace-guards`** to log each transition guard result.

## Library

Stable imports for embedding the engine in your own service or tool:

- **`github.com/justinush/maestro/pkg/definition`** — workflow types and **`DecodeFile`** (strict YAML/JSON).
- **`github.com/justinush/maestro/pkg/engine`** — **`NewInstance`**, **`RunUntilBlocked`**, **`SubmitInput`**, **`Registry`**, **`ActionRunner`**, trace **`Event`** values, and sentinel errors.
- **`github.com/justinush/maestro/pkg/validate`** — **`WorkflowFile`** and **`WorkflowDefinition`** (same checks as **`maestro validate`**).
- **`github.com/justinush/maestro/pkg/run`** — **`Store`**, **`MemoryStore`**, and **`RunRecord`** for persisting **`engine.Snapshot`** (optimistic **`revision`**); pair with **`engine.NewInstanceFromSnapshot`** after reloading the same workflow.

Typical flow: **`definition.DecodeFile`** → optional **`validate.WorkflowDefinition`** → **`engine.NewInstance`** → **`RunUntilBlocked`** / **`SubmitInput`** (then **`Snapshot`** / **`pkg/run`** when you need durability).

Minimal program:

```bash
go run ./examples/library examples/workflow-v0-minimal.yaml
```

After you publish tags, pin with **`require github.com/justinush/maestro v0.x.y`** in the consumer module.

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

## Action runners

Actions on a step (`onEnter` / `onExit`) are dispatched by **type** through a **`Registry`**:

| Runner | `type` in YAML | Role |
|--------|----------------|------|
| **Stub** | `stub` | Set **`variables`** from JSON (`params.set`); no I/O. |
| **HTTP** | `http` | Outbound request; writes result under **`resultVariable`** (`statusCode`, `headers`, `body`). |
| **Custom** | your string | Implement **`ActionRunner`** and **`Register`** it (same pattern as built-ins). |

The default registry used when **`ActionRegistry`** is omitted is **stub-only**. **`simulate`** registers **stub + HTTP** for scenarios that need it.

## Persistence and execution model

The **`engine.Instance`** holds **live** state: current step, **`variables`**, partial trace, and compiled guard caches. Nothing is written to disk unless **your application** does so.

For durability Maestro exposes:

- **`Instance.Snapshot()`** — JSON-friendly state (`currentStepId`, **`variables`**, **`onEnterRan`**, trace events, sequence counter). CEL and input-schema caches are **not** in the snapshot; they are rebuilt when you restore.
- **`engine.NewInstanceFromSnapshot(def, snap, opts)`** — rebuild an **`Instance`** from the same **workflow definition** plus a snapshot (same **`def.id` / `def.version`** and graph you used when the run started).

**`pkg/run`** adds a small **`Store`** interface and **`MemoryStore`** (process-local, optimistic **`revision`** for concurrent updates). Implement **`Store`** with Postgres or another backend when you need shared or durable runs.

## Current scope and non-goals

### In scope today

- **Synchronous** step execution: **`RunUntilBlocked`** advances **`action`** steps until blocked or completed; **`human`** steps block until **`SubmitInput`**.
- **Declarative** workflows (v0.1 schema), **CEL** exit guards, **JSON Schema** for human input.
- **Embeddable** Go API (`pkg/definition`, `pkg/engine`, `pkg/validate`, `pkg/run`).
- **CLI**: **`validate`**, **`simulate`** (assertions, trace).
- **Tracing** in memory for the lifetime of the instance.

### Non-goals (for now)

- **Distributed** workflow execution across nodes (single-instance **`Instance`** model).
- Built-in **retry policies**, **cron**, or **job scheduler** (your orchestrator owns that).
- **Visual workflow designer** or hosted multi-tenant SaaS.
- **Automatic** pre-resolution of **external JSON Schema `$ref`** chains (custom **`$id`** is supported; deep **`$ref`** loading is not).

## Roadmap

Near-term directions (not commitments):

- **Clearer stop / completion API** for embedders (e.g. explicit stop reason alongside **`error`**).
- **`context.Context`** through **`RunUntilBlocked`** and **`ActionRunner`** for cancellation and deadlines.
- **Persistence adapters** beyond **`MemoryStore`** (documented examples for SQL).
- **Trace**: timestamps, optional persistence, stable JSON contract for exporters.
- **CEL**: optional cost limits; shared-env optimizations if profiles show need.
- **JSON Schema**: optional **`$ref`** loader for shared schema artifacts.
- **Scenario / contract tests**: more packaged journeys like **`examples/kyc/sg/portrait`**.

## Contributing

- Run **`make check`** (lint, vet, tests) before opening a PR.
- **`make smoke`** exercises validate + simulate paths used in CI-style checks.
- Prefer small, focused changes with tests when behavior shifts.

Suggestions and bug reports are welcome via issues and pull requests on the repository.

## License

Maestro is released under the **MIT License** (see [`LICENSE`](LICENSE)).
