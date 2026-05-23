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
    ↓
Runtime
    ↓
Instance
    ↓
RunUntilBlocked()
```

Minimal example:

```go
rt, err := maestro.Load("workflow.yaml")

in, err := rt.NewInstance(maestro.InstanceOptions{})

res := in.RunUntilBlocked()
```

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
    ↓
RunUntilBlocked
    ↓
persist run
    ↓
future request
    ↓
restore run
    ↓
SubmitInput
    ↓
continue execution
```

This model fits naturally into:
- HTTP APIs
- approval systems
- onboarding flows
- async operational workflows

---

# Persistence Model

Workflow state is stored separately from your business data.

Maestro persists:
- current step
- variables
- execution state
- execution trace

Your application persists:
- users
- applicants
- documents
- business entities

Example:

```go
rec := run.RecordFromInstance(in, def, revision)
```

Restore later:

```go
in, err := rt.RestoreInstance(rec, maestro.InstanceOptions{})
```

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

Maestro currently focuses on:

- human-in-the-loop workflows
- backend orchestration
- embedded execution
- persistence + restore
- request-driven lifecycle management

The examples in this repository demonstrate:
- minimal embedding
- persistence flows
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
