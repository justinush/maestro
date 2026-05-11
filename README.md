# Maestro

Extensible workflow runtime for KYC and onboarding systems.

## The problem

Onboarding flows have a habit of leaking into application code: another `if` for a new corridor, vendor response handling wedged next to routing logic, and three services that all “know” what step comes next. It works until it doesn’t.

You end up with:

- Journeys that are hard to read as a whole, and painful to diff or review.
- Transitions that only exist inside services, so reuse across markets or entities is awkward.
- Vendor-specific paths tangled with orchestration, so swapping or testing integrations is expensive.
- Debugging and audits that depend on piecing together logs instead of a clear run story.
- Every regional tweak turning into a deploy and a round of regression testing.

None of that is novel—it’s just the cost of encoding workflows by hand once the product gets broad enough.

## The solution

Maestro splits the problem in two: **what the journey is** (a file checked into git) and **how you run it** (a Go library you embed wherever you already run Go).

You get:

- YAML/JSON definitions with steps, transitions, and CEL guards—fewer ad hoc `switch` trees for “what’s next.”
- The same engine behind the CLI and your service or worker tests.
- Action types (`stub`, `http` today) behind a registry, plus room for your own runners.
- `validate` and `simulate` so bad definitions fail early and scenarios can assert on outcomes.
- A trace of what happened during a run while it lives in memory.
- Snapshots plus a tiny `pkg/run` store interface (in-memory implementation included) when you’re ready to persist without dragging a database into the core.

Maestro is not intended to replace application code or business systems. It focuses on workflow orchestration primitives and embeddable runtime execution.

## Why Maestro?

Most teams start with plain code: `if` chains, service-local state, vendor SDKs in the same package. That’s fine at small scale. It gets brittle when corridors multiply and every change feels like a production gamble.

Roughly:

- **Today:** “what happens next” is implied by scattered code; hard to see the journey as one artifact.
- **With Maestro:** the journey is a definition you can read top to bottom, version, and argue about in review.

- **Today:** onboarding rules and HTTP calls to vendors often share a file.
- **With Maestro:** orchestration (steps, guards, `variables`) stays separate from runners that do I/O.

- **Today:** another country or entity often means another branch in code.
- **With Maestro:** often another definition (or the same graph with different data in `variables`) instead of rewriting the engine.

- **Today:** “what did this user hit?” is reconstructed from logs.
- **With Maestro:** you get a structured trace; `simulate` can pin branches in CI so regressions show up before users do.

You still write integration code. The point is to stop burying the whole story inside it.

## Core concepts (v0.1)

**Workflow definition** — YAML/JSON with `steps` and `transitions`. Schema version is `0.1` for now.

**Step kinds**

- `human` — waits for `SubmitInput`.
- `action` — runs `onEnter` / `onExit` actions, then moves on via transitions.
- `end` — terminal; must appear in `terminalStepIds`.

### Exit guards

To leave a step you pick a **transition**. If `transitions[].when` is set, it’s a CEL **exit guard** (must be true for that edge). Empty `when` means unconditional, with ordering by `priority` then declaration order when ties matter.

Practical habit: let `onEnter` / `onExit` and `SubmitInput` write the fields your guards read in `variables`, instead of hiding state only in UI or opaque blobs.

Guards see `variables` in CEL today—keep names stable between actions, guards, and tests.

### Variables

`variables` is the bag of instance state. Top-level keys come from `stub`, from `http` (via `resultVariable`), and from a shallow merge on `SubmitInput`. Nested values are fine; you’re still replacing whole top-level keys unless you model paths yourself.

Suggested keys (the engine doesn’t enforce this—it’s convention):

| Area | Key | Notes |
|------|-----|--------|
| Partner / sync | `partner` | e.g. `status`, errors |
| Base / profile | `profile` or flat keys | whatever you agree at “COR-style” steps |
| Liveness | `liveness` | e.g. `passed` |
| Address / POA | `address` | e.g. `status`, `reviewRequired` |
| Integrations | `integration`, … | probe payloads, flags |

Use the same names in definitions, `simulate`’s `initialVariables`, and `expectVariables` so CEL and scenarios stay boring in a good way.

### Input schema

`steps[].inputSchema` is JSON Schema. On `human` steps, `SubmitInput` is checked against it before merge.

## Architecture

Maestro is meant to be **imported**: definitions and checks sit on top of a small engine; side effects go through runners; your app owns databases, auth, queues, and whatever stores snapshots.

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
       │   Instance   │  current step, variables, guards,
       │   (engine)   │  RunUntilBlocked / SubmitInput
       └──────────────┘
              │
              ▼
       ┌──────────────┐
       │Action runners│  Registry: stub, http, yours
       └──────────────┘
              │
              ▼
  Your app (HTTP, queues, pkg/run Store, UI)
```

The `maestro` binary is thin glue over `pkg/validate` and `pkg/engine`—same code paths you’d use in production.

## Requirements

- [Go](https://go.dev/dl/) **1.26+** (see `go.mod`)

## Quick start

```bash
go build -o maestro ./cmd/maestro
./maestro validate -f examples/workflow-v0-minimal.yaml
```

Without installing a binary:

```bash
go run ./cmd/maestro validate -f examples/workflow-v0-minimal.yaml
```

Run a scripted scenario (with assertions):

```bash
go run ./cmd/maestro simulate -s examples/scenario-minimal.yaml
```

Trace (plain text or JSON):

```bash
go run ./cmd/maestro simulate -s examples/scenario-minimal.yaml --trace
go run ./cmd/maestro simulate -s examples/scenario-minimal.yaml --trace --trace-format json
go run ./cmd/maestro simulate -s examples/scenario-minimal.yaml --trace-guards
```

Custom schema and louder errors:

```bash
./maestro validate -f path/to/workflow.yaml --schema path/to/schema.json --verbose
```

Smoke the minimal example:

```bash
make validate-example
```

Broader smoke (minimal + negative scenarios + Singapore portrait):

```bash
make smoke
```

## Examples

### Minimal onboarding

See `examples/workflow-v0-minimal.yaml`: human `collect-profile` → action `run-checks` (stub flips `checksStarted`) → `done`, with CEL on the edge into `done`. Drive it with `examples/scenario-minimal.yaml`.

### Singapore portrait

Under `examples/kyc/sg/portrait/` there’s a longer path (partner gate, relationship capture, stubbed integration, liveness, address, several end states). Scenarios: `scenario-happy.yaml`, `scenario-partner-rejected.yaml`, `scenario-poa-review.yaml`. The checked-in workflow uses stubs for the integration probe so `make smoke` stays off the public internet.

### What a trace looks like

```bash
go run ./cmd/maestro simulate -s examples/scenario-minimal.yaml --trace
```

You’ll see lines like (ids may differ if you change the example):

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

`--trace-format json` for machines; `--trace-guards` if you’re debugging why a guard didn’t fire.

## Library

Import paths (same module you’d `go get` after tagging):

- `github.com/justinush/maestro/pkg/definition` — types and strict `DecodeFile`.
- `github.com/justinush/maestro/pkg/engine` — `NewInstance`, `RunUntilBlocked`, `SubmitInput`, registry, `ActionRunner`, events, errors, snapshots.
- `github.com/justinush/maestro/pkg/validate` — same checks as `maestro validate`.
- `github.com/justinush/maestro/pkg/run` — `Store`, `MemoryStore`, `RunRecord` around `engine.Snapshot` with optimistic `revision`; reload with `NewInstanceFromSnapshot` and the **same** workflow definition.

Happy path: decode → optionally validate → `NewInstance` → loop `RunUntilBlocked` / `SubmitInput` → snapshot to a store when you care about survival past the process.

Tiny embed example:

```bash
go run ./examples/library examples/workflow-v0-minimal.yaml
```

That one stops at the first human step on purpose—it’s a skeleton, not a full product.

When you publish versions, pin in the consumer’s `go.mod`, e.g. `require github.com/justinush/maestro v0.1.0`.

## CLI

### `maestro validate`

Strict decode, JSON Schema (embedded v0.1 or `--schema`), graph checks, CEL compile on `when`, stub/http params, `inputSchema` compile.

### `maestro simulate`

Loads a scenario file: initial `variables`, scripted inputs, assertions on completion or errors, optional trace flags.

## Development

```bash
make test          # or: go test ./...
make check         # lint + vet + test
make build         # binary under dist/
make smoke         # validate + simulate (minimal + negative + portrait)
```

`make help` lists the rest.

## Action runners

`onEnter` / `onExit` actions are keyed by `type` in the registry:

| | `type` | Does |
|---|--------|------|
| Stub | `stub` | Writes `params.set` into `variables`; no network. |
| HTTP | `http` | One request; result under `resultVariable` (`statusCode`, `headers`, `body`). |
| Yours | anything | Implement `ActionRunner`, `Register` it like the built-ins. |

If you don’t pass `ActionRegistry`, you get stub only. The `simulate` command wires stub + HTTP so examples can hit real HTTP when the workflow asks for it.

## Persistence and execution model

An `Instance` is memory-only: current step, `variables`, trace, compiled guards. Nothing hits disk unless you write it.

When you’re ready:

- `Snapshot()` captures JSON-friendly state (step id, `variables`, `onEnterRan`, events, seq). CEL / input-schema caches are rebuilt on restore—not stored.
- `NewInstanceFromSnapshot(def, snap, opts)` rebuilds from the same definition artifact you used when the run started (same id/version and graph).

`pkg/run` adds `Store` and a `MemoryStore` with a simple revision counter for “last write wins” style conflicts. Swap in Postgres or anything else behind `Store` when you need shared or durable runs.

## Current scope and non-goals

**In scope today** — synchronous stepping: `RunUntilBlocked` walks `action` steps until it blocks or finishes; `human` waits on `SubmitInput`. v0.1 schema, CEL guards, JSON Schema on human input. CLI `validate` / `simulate`. Trace for the life of the instance. Go packages above.

**Not trying to be (yet)** — a distributed workflow cluster, a built-in scheduler or retry engine, a drag-and-drop designer, or a hosted multi-tenant product. We also don’t auto-fetch arbitrary external JSON Schema `$ref` chains (custom `$id` is fine; deep ref loading isn’t).

## Roadmap

Ideas we care about but haven’t promised on a calendar:

- Clearer “stopped because …” surface for embedders than sentinel errors alone.
- Threading `context.Context` through the run loop and runners.
- Example `Store` implementations people can copy for SQL.
- Trace timestamps, optional persistence, stable JSON for exporters.
- CEL guard budgets if real workloads need them.
- Better `$ref` story for shared schemas.
- More scenario packs like the portrait folder.

## Contributing

Run `make check` before a PR; `make smoke` is the closest thing to “what CI cares about” locally. Small diffs with tests when behavior changes beat large refactors.

Issues and PRs are welcome.

## License

MIT — see [`LICENSE`](LICENSE).
