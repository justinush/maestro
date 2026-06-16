# Architecture

This document explains the core execution model behind Maestro.

Maestro is an embedded workflow orchestration runtime originally designed for KYC and onboarding systems.

It helps backend applications handle flows involving:

- human approval steps
- external vendor actions
- pause/resume execution
- persistence + restore
- workflow-driven orchestration

Unlike orchestration platforms that require separate infrastructure, Maestro runs directly inside your Go application as a library.

---

# High-Level Model

Your application owns the business system.

Maestro owns workflow orchestration.

![Architecture](./assets/architecture.png)

Typical ownership split:

| Your Application | Maestro |
|---|---|
| HTTP APIs | workflow execution |
| authentication | transitions |
| database models | human steps |
| UI | action orchestration |
| business logic | pause/resume lifecycle |
| vendor integrations | execution trace |

---

# Runtime Lifecycle

A workflow definition becomes a runtime.

A runtime creates workflow instances.

```txt
workflow.yaml
    -> Runtime
    -> Instance
    -> RunUntilBlocked()
```

Minimal example:

```go
rt, err := maestro.Load("workflow.yaml")

in, err := rt.NewInstance(maestro.InstanceOptions{})

res := in.RunUntilBlocked()
```

---

# Public API

**Canonical embed path:** `pkg/maestro`

```txt
maestro.Load*
    -> Runtime.NewInstance
    -> RunUntilBlocked
    -> SubmitInput (when blocked, repeat)
```

**Persistence:** `pkg/run` (`Store`, `RunRecord`, `RecordFromInstance`)

```txt
RecordFromInstance -> Store.Create / Save
Store.Get -> rt.RestoreInstance   (preferred in apps)
```

**Postgres store:** `pkg/run/postgres` (`NewStore`, `SchemaDDL`) implements `run.Store` with Postgres + JSONB. Optional `ApplySchema` for demos and integration tests; production apps typically apply the same DDL via their migration tool. `run.NewMemoryStore` is for tests and demos only.

**Advanced:** `pkg/engine` for direct `Options`, registries, and `NewInstanceFromSnapshot`; `run.InstanceFromRecord` for store-layer restore without a `Runtime`.

**HTTP actions:** pass your own `*http.Client` to `engine.RegistryWithHTTP`. The `maestro simulate` CLI uses a private long-timeout client (not exported from `pkg/engine`).

## Workflow registry

Multi-workflow apps:

For **single-workflow** apps, `maestro.Load` returns one `Runtime` — the path above is enough.

For **multi-workflow** apps, `pkg/workflow` provides a registry that maps workflow **`id` + `version`** to a validated `Runtime` (`LoadDir`, `Register`, `Lookup`, `NewInstance`, `RestoreInstance`).

The registry is intentionally **not** a business router. Applications still own routing such as:

```txt
country + flow type  ->  workflow id + version   (host map or config)
```

Maestro owns:

```txt
workflow id + version  ->  Runtime  ->  Instance
```

`RunRecord` already stores `workflowId` and `workflowVersion`; `Registry.RestoreInstance` looks up the correct definition for an existing run. Single-workflow apps can keep using `rt.RestoreInstance` unchanged.

Which version **new** runs should use (rollouts, rollbacks, “active” version) is host policy — see [Workflow versioning design](design/workflow-versioning.md).

---

# Workflow Definitions

Workflows are declared in YAML or JSON.

Example:

```yaml
schemaVersion: "0.1"

id: onboarding
version: "1.0.0"

initialStepId: collect-profile

terminalStepIds:
  - approved

steps:
  - id: collect-profile
    kind: human

  - id: approved
    kind: end

transitions:
  - from: collect-profile
    to: approved
```

A workflow contains:

- steps
- transitions
- guards
- input schemas
- actions

---

# Human Steps

Human steps pause execution until input is submitted.

```yaml
- id: collect-profile
  kind: human
```

When execution reaches a human step:

```go
res := in.RunUntilBlocked()
```

The engine returns:

```go
engine.RunBlocked
```

Your application can then:
- render UI
- wait for API requests
- persist the workflow run

Later:

```go
sub := in.SubmitInput(...)
```

The workflow continues from the blocked step.

---

# Action Steps

Action steps execute registered runners.

Example:

```yaml
- id: run-liveness-check
  kind: action
```

Actions usually represent:
- vendor APIs
- internal services
- webhooks
- queues
- verification systems

Runners are registered through an action registry.

Example:

```go
reg := engine.RegistryWithHTTP(client)
```

---

<a id="pause-resume-model"></a>

# Pause / Resume Model

One of Maestro's core ideas is resumable orchestration.

A workflow run can be:

- executed
- paused
- stored
- restored
- continued later

Typical lifecycle:

```txt
start request
    -> RunUntilBlocked
    -> persist run
    -> future request
    -> restore run
    -> SubmitInput
    -> continue execution
```

This model fits naturally into:
- HTTP APIs
- approval systems
- onboarding flows
- async operational workflows

Multi-day waits (webhooks, approval links, provider callbacks) use the same persist → restore → resume loop — see [Async callbacks design](design/async-callbacks.md). The bridge pattern (action create + resume step + host `externalRef` table) is validated; hosts should not hold HTTP requests open until a callback arrives.

---

# Persistence Model

Workflow state is stored separately from your business data.

Maestro persists (in `run.RunRecord` / `engine.Snapshot`):
- current step
- variables
- execution state
- execution trace

Your application persists:
- users
- applicants
- documents
- business entities

## Revision and optimistic locking

`RunRecord.Revision` enables optimistic concurrency on `Store.Save`.

| Step | Revision behavior |
|------|-------------------|
| `RecordFromInstance(..., 0)` + `Store.Create` | Stored revision becomes **1** |
| `Store.Get` | Returns the current revision |
| `Store.Save` | Succeeds only if `rec.Revision` matches; then revision increments |
| Concurrent `Save` | Loser gets `run.ErrRevisionConflict` |

JSON tags on `RunRecord` and `Snapshot` (`runId`, `revision`, `currentStepId`, …) are intended to remain stable across v0.x releases.

## Save and restore

Build a record after a request mutates the instance:

```go
rec := run.RecordFromInstance(in, def, 0)
err = store.Create(ctx, rec)
```

Later request:

```go
rec, err := store.Get(ctx, runID)
in, err := rt.RestoreInstance(rec, maestro.InstanceOptions{
	ActionRegistry: reg,
})
```

`rt.RestoreInstance` is the preferred application API. `run.InstanceFromRecord` is the lower-level equivalent when implementing or calling `Store` directly.

Continue execution with `SubmitInput` and `RunUntilBlocked` as on the first request.

## Postgres adapter

`pkg/run/postgres` is the reference implementation of `run.Store`.

| Piece | Role |
|-------|------|
| `schema.sql` / `SchemaDDL` | Canonical `workflow_runs` DDL (stable across v0.x, intended) |
| `ApplySchema` | Optional idempotent apply — demos, integration tests, local dev |
| `NewStore(pool)` | `Create` / `Get` / `Save` with JSONB state and revision-based optimistic locking |
| `workflow_runs` table | `run_id` PK, `revision`, `state` JSONB |

Setup (schema must exist before `NewStore`):

```go
pool, err := pgxpool.New(ctx, databaseURL)
if err != nil { ... }

// Option A (optional): quick setup for demos/tests
// err = postgres.ApplySchema(ctx, pool)

// Option B (typical in production): migrations from postgres.SchemaDDL() or schema.sql

store := postgres.NewStore(pool)
```

**Schema management:** `ApplySchema` is a convenience, not a requirement. Production services usually version DDL with their own migration tool and treat `schema.sql` as the source of truth.

The adapter persists **workflow runs only**. Application tables (users, applicants, documents, etc.) remain your responsibility. Maestro does not coordinate a single database transaction across `run.Store` and your business tables in v0.1; production services should define their own transaction boundaries.

Table and column names in `pkg/run/postgres/schema.sql` are intended to remain stable across v0.x releases, alongside `RunRecord` JSON field names.

---

# Variables

Workflow variables carry orchestration state across steps.

Example:

```yaml
when: "variables.review.required == true"
```

Variables are commonly used for:
- branching
- guard conditions
- action outputs
- review decisions
- orchestration metadata

---

# Execution Trace

Each workflow instance records execution events.

Example:

```go
events := in.Events()
```

This helps with:
- debugging
- support tooling
- operational visibility
- audit-style inspection

The `kyc-backend` example exposes these through:

```txt
GET /kyc/{runID}/events
```

---

# API stability (v0.x)

Treat as **stable for v0.x** (breaking changes -> v1):

- `RunRecord` / `Snapshot` / `Event` JSON field names
- `RunStatus` and `SubmitInputStatus` enum values
- Sentinel errors such as `run.ErrNotFound`, `run.ErrRevisionConflict`
- `workflow_runs` table/column names in `pkg/run/postgres` (intended stable across v0.x)

`pkg/maestro` is the supported embedding surface. Symbols in `pkg/engine` and `pkg/run` remain public for advanced use but may gain helpers without a major bump; semantic changes to persistence JSON or status enums will not.

---

# Embedded-First Philosophy

Maestro intentionally avoids requiring:
- workflow servers
- orchestration clusters
- external schedulers
- dedicated infrastructure

Instead:

```txt
your application
    embeds
Maestro directly
```

This keeps orchestration close to:
- business logic
- APIs
- persistence
- operational ownership

---

# Current Scope

**v0.1.0** ships the embedded runtime, `run.Store` (memory + Postgres), and CLI tooling. See [roadmap](./roadmap.md) for what's next.

Maestro focuses on:

- human-in-the-loop workflows
- backend orchestration
- embedded execution
- persistence + restore
- request-driven lifecycle management

The examples in this repository demonstrate:
- minimal embedding
- persistence flows (in-memory; Postgres lives in `pkg/run/postgres`)
- HTTP integrations
- realistic KYC orchestration

---

# Related Examples

| Example | Purpose |
|---|---|
| [`library-basic`](../examples/demos/library-basic) | smallest embedding example |
| [`embed-kyc-service`](../examples/demos/embed-kyc-service) | pause/resume + persistence |
| [`http-runner`](../examples/demos/http-runner) | HTTP action integration |
| [`kyc-backend`](../examples/apps/kyc-backend) | realistic backend service |
