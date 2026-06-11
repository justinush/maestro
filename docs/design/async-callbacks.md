# Async callbacks design (v0.2.x)

How to resume a workflow days later — when a webhook fires, a manager clicks a link, or a payment provider calls back — without holding an HTTP request open.

This doc covers **waiting and resume**: what Maestro does today, what production hosts need, and what we may add to the engine next. Persistence and restore mechanics are in [architecture — pause / resume](../architecture.md#pause-resume-model). Workflow identity on restore follows [workflow versioning](workflow-versioning.md).

**Status:** design draft. No new APIs are committed until we agree on this text and try the bridge pattern in a real host.

---

## Goals

- Support **multi-day** flows: vendor verification, manager approval, payment callbacks.
- Keep the embed model: the host owns HTTP routes, webhooks, and vendor SDKs — Maestro owns blocked snapshot + resume semantics.
- Build on what already works (`RunBlocked`, `SubmitInput`, `RunRecord`, restore).
- Make the **host resume recipe** obvious: restore → deliver payload → `RunUntilBlocked` → save.

## Non-goals (for this pass)

- A built-in inbound webhook server or callback router product.
- Polling a vendor for days inside the engine (that belongs with retries/timers — separate roadmap item).
- Storing vendor secrets in workflow YAML.
- A distributed worker cluster or queue as a Maestro built-in.
- Replacing the host’s idempotency, auth, and vendor-id mapping layers.

---

## The problem

Today the happy path looks synchronous:

```text
human step → RunUntilBlocked → RunBlocked
           → SubmitInput → RunUntilBlocked → …
```

Real systems look like this:

```text
Create vendor verification (HTTP)
    → wait hours or days
    → vendor webhook arrives
    → resume workflow

Send approval email
    → manager clicks link tomorrow
    → resume workflow

Call payment provider
    → provider callback
    → resume workflow
```

Action steps today run **to completion inside one** `RunUntilBlocked` call. The HTTP action runner blocks until the response returns or times out. There is no “fire request, pause the run, resume on webhook” inside a single action step.

That gap is what this design addresses.

---

## What already works

**Delayed human input** is already async.

When execution hits a `kind: human` step, `RunUntilBlocked` returns `RunBlocked`. The host persists `RunRecord` and returns. Hours or weeks later:

```go
rec, _ := store.Get(ctx, runID)
in, _ := rt.RestoreInstance(rec, maestro.InstanceOptions{ActionRegistry: reg})

sub := in.SubmitInput(payload)          // validate, merge variables, advance if a transition fires
_ = in.RunUntilBlocked()                // drive action steps until next block or terminal
_ = store.Save(ctx, run.RecordFromInstance(in, def, rec.Revision))
```

The [kyc-backend demo](../../examples/apps/kyc-backend/) follows this pattern: `Start` blocks on the first human step; later API calls restore and `SubmitInput` on the expected step.

**Correlation id:** `RunID` on the instance and in `RunRecord` / `Snapshot`. That is the primary key hosts use to tie a callback to a run.

**Concurrency:** `RunRecord.Revision` + `Store.Save` optimistic locking — two concurrent resumes must not silently clobber each other.

So: **approval-link tomorrow** and **form submit next week** are not missing engine features. They need host HTTP handlers and persistence discipline, not a new primitive — unless we want clearer YAML for non-UI waits (see [Bridge pattern](#bridge-pattern-wait-without-a-new-primitive) below).

---

## The gap

| Need | Today |
|------|--------|
| Pause after outbound vendor call, resume on webhook | Action steps do not yield mid-step; HTTP runner is synchronous |
| Express “wait for external event” in YAML without pretending it is a UI form | Only `kind: human` blocks |
| Webhook payload → variables → transition | Works **if** the run is blocked on a step that accepts `SubmitInput` |

Vendor-shaped async is the main gap. Human-shaped async already works.

---

## Waiting state (engine model)

A run is **waiting** when the host must not call `RunUntilBlocked` again until it delivers external input.

| Field | Meaning |
|-------|---------|
| `RunResult.Status == RunBlocked` | Engine yielded on a human step |
| `Snapshot.CurrentStepID` | Step that must receive the next resume |
| `Snapshot.OnEnterRan` | Whether `onEnter` actions on that step already ran |
| `Snapshot.Variables` | State merged from prior steps and prior inputs |
| `RunRecord.WorkflowID` / `WorkflowVersion` | Definition used on restore (exact version — see versioning doc) |

Trace event `run.blocked` (`EventBlocked`) is recorded when blocking.

**Contract:** every host request that ends in a wait must **persist** the snapshot before returning. Never keep a process open until a webhook arrives.

### Waiting kinds (conceptual)

| Kind | Who delivers resume | Engine API today |
|------|---------------------|------------------|
| **Human** | End user, backoffice API, approval link handler | `SubmitInput` on `kind: human` |
| **External / vendor** | Webhook handler after vendor event | Same as human if modeled as a wait step (bridge); dedicated primitive TBD |

We may later distinguish external waits in status or step kind. Until then, the bridge uses human blocking semantics under the hood.

---

## Callback identity

Maestro does not receive webhooks. The host does. Correlation is split on purpose.

### What Maestro stores

```text
runId              — primary key on RunRecord
currentStepId      — in Snapshot; which step expects resume
workflowId + version — on RunRecord for RestoreInstance
```

### What the host adds

```text
Vendor id → runId mapping     (e.g. verification session id in Postgres)
Signed token in callback URL  (runId + stepId + expiry + HMAC)
Webhook signature verification
Idempotency on vendor event id (safe to process duplicate delivery)
```

**Minimum resume key:** `(runId, stepId)` — after restore, the host checks `in.CurrentStepID()` matches the step the callback is for. Wrong step → reject (see kyc-backend `ErrWrongStep` pattern).

Maestro does not plan a global callback registry in core. Optional host tables and URL shapes are integration details.

---

## Host resume recipe

Canonical handler for any async completion (approval, webhook, provider callback):

```go
// 1. Resolve run (from URL token, body, or vendor-id lookup)
rec, err := store.Get(ctx, runID)

// 2. Restore exact workflow version
in, err := reg.RestoreInstance(rec, maestro.InstanceOptions{ActionRegistry: reg})

// 3. Guard: callback is for the step we think
if in.CurrentStepID() != wantStep {
    return errWrongStep
}

// 4. Deliver payload (today: SubmitInput on human / wait step)
sub := in.SubmitInput(normalizedPayload)
if sub.Status == engine.SubmitFailed {
    return sub.Err
}

// 5. Drive automatic work (action steps) until next block or done
res := in.RunUntilBlocked()
switch res.Status {
case engine.RunBlocked, engine.RunCompleted:
    // ok
case engine.RunFailed:
    return res.Err
}

// 6. Persist with revision check
return store.Save(ctx, run.RecordFromInstance(in, def, rec.Revision))
```

**Host owns:** steps 1 and 3 (auth, mapping, idempotency, normalizing vendor JSON to the input map).

**Maestro owns:** steps 4–6 semantics (validation against `inputSchema`, transitions, action execution, snapshot).

**Idempotency:** if the vendor redelivers the same event, the host should detect “already advanced past this step” or “duplicate event id” before calling `SubmitInput` again. Engine-level dedup is not in scope for the first slice.

---

## Maestro vs host

| Maestro | Host |
|---------|------|
| `RunBlocked` / `RunCompleted` / `RunFailed` | HTTP routes for webhooks and approval links |
| Snapshot + `RunRecord` shape | Vendor SDKs and outbound API calls |
| `SubmitInput` (+ future external resume API if we add one) | Vendor payload → input map |
| `inputSchema` validation on resume | Webhook auth, replay protection |
| Execution trace (`run.blocked`, `input.accepted`, …) | When to call vendor (which step’s `onEnter`) |
| Restore by exact `(workflowId, version)` | Retry polling, DLQ, alerting |

---

## Scenario patterns

### Manager approval (link clicked tomorrow)

| | |
|--|--|
| **Graph** | `kind: human` step with `inputSchema` for `{ approved: bool }` |
| **Wait** | `RunUntilBlocked` → `RunBlocked` → persist |
| **Resume** | Link hits host → restore → `SubmitInput({ approved: true })` → `RunUntilBlocked` |
| **Engine change** | None required |

This is the same model as a backoffice API — only the transport differs.

### Vendor verification (create session, webhook days later)

| | |
|--|--|
| **Graph (bridge)** | `action` step: `onEnter` HTTP POST creates session, stores vendor id in `variables` → `wait-vendor` human step (minimal schema) → transitions on webhook payload |
| **Wait** | Blocked on `wait-vendor` |
| **Resume** | Webhook handler looks up `runId` from vendor id → `SubmitInput` with normalized result → `RunUntilBlocked` |
| **Engine change** | None for bridge; optional `kind: wait` later for clearer YAML |

### Payment provider callback

Same as vendor verification: outbound action (or host code before block) + wait step + webhook → `SubmitInput`. Provider signature verification and idempotency stay in the host.

---

## Bridge pattern (wait without a new primitive)

Until we add a dedicated external-wait step or signal API, model vendor waits as a **human step used only as a pause point**:

```yaml
- id: create-verification
  kind: action
  onEnter:
    - type: http
      id: create-session
      params:
        method: POST
        url: "https://vendor.example/v1/sessions"
        resultVariable: vendor

- id: wait-vendor-result
  kind: human
  description: Blocks until vendor webhook; no end-user UI.
  inputSchema:
    type: object
    required: [status]
    properties:
      status:
        type: string
      vendorRef:
        type: string

transitions:
  - from: create-verification
    to: wait-vendor-result
  - from: wait-vendor-result
    to: next-step
    when: "variables.status == 'approved'"
```

**Pros:** ships on today’s engine; explicit in the graph; `inputSchema` documents the webhook payload.

**Cons:** `kind: human` for non-human events is a modeling compromise; `presentationRef` may be empty or a host convention.

We treat this as an **official bridge**, not a hack — documented here until a clearer step kind or signal API exists.

---

## Design options (engine)

If the bridge is not enough, pick **one** primary primitive in a follow-up implementation pass — not all at once.

| Option | Idea | Pros | Cons |
|--------|------|------|------|
| **Bridge only** | Human wait step | Zero engine change | YAML semantics |
| **`kind: wait`** | New step kind; blocks like human; `SubmitInput` or `DeliverCallback` | Clear graph | Schema + engine loop change |
| **`DeliverSignal(name, payload)`** | Signal while blocked; name matches step or subscription | Webhook-shaped; UI stays on `SubmitInput` | New API; overlap rules with human |
| **Async action runner** | Action returns pending; new `RunStatus` | Matches “one step” mentally | Harder to test; couples with retries |

**Current lean:** ship the **bridge** pattern first (docs + optional `examples/`); add **`kind: wait`** or **`DeliverSignal`** only if host integration proves the bridge is too awkward. Avoid async-only action magic in the first engine change.

Relationship to **retries / timers** (roadmap): callbacks are **push** (event arrives). Retries are **pull** (engine or host tries again). They compose but ship as separate designs.

---

## Relationship to shipped APIs

Already available:

- `Instance.RunUntilBlocked()` → `RunBlocked` on human steps
- `Instance.SubmitInput(map[string]any)` — human steps only
- `run.RecordFromInstance` / `Store.Save` / `RestoreInstance`
- `workflow.Registry.RestoreInstance` for multi-workflow hosts
- `RunID` on instance and snapshot

Not available today:

- Pause mid-action-step
- `DeliverSignal` / callback-specific resume
- `kind: wait` in workflow YAML
- Engine-side webhook idempotency store

---

## Scope

| Phase | Deliverable | Notes |
|-------|-------------|--------|
| **1 — design** | This doc + architecture / roadmap links | Agree on rules before code |
| **2 — bridge** | Optional `examples/` snippet | Vendor create + human wait step; **no engine change required** |
| **3 — engine (if needed)** | One resume primitive | `kind: wait` **or** `DeliverSignal` — not both in first implementation; CLI simulate support for webhook inject |

Explicitly **not** in scope: inbound webhook server, vendor-specific adapters, callback DB inside Maestro.

---

## Open questions

1. **Native primitive name** — `wait` step vs `signal` vs generalized `Resume(stepID, payload)`?
2. **Bridge deprecation** — when native wait ships, do we keep recommending human-as-wait for simple cases?
3. **`inputSchema` on wait steps** — required for vendor payloads, or optional with host validation only?
4. **Status subtyping** — does `RunBlocked` need a reason (`human` vs `external`), or is `step.kind` after restore enough?
5. **Simulate tooling** — `maestro simulate` inject callback step as `submit` vs new `signal` command?

---

## Related docs

- [Architecture — pause / resume](../architecture.md#pause-resume-model)
- [Architecture — persistence](../architecture.md#persistence-model)
- [Workflow versioning](workflow-versioning.md) — exact version on restore; long-running runs
- [Roadmap](../roadmap.md)
- [kyc-backend demo](../../examples/apps/kyc-backend/) — restore + `SubmitInput` reference host

---

## Summary

- **Human async already works:** `RunBlocked` → persist → later `SubmitInput` → `RunUntilBlocked` → save.
- **Vendor async needs a pattern:** bridge today (`action` create + `human` wait step + webhook handler); optional engine primitive later.
- **Callback identity:** `runId` + `stepId` in Maestro; vendor mapping and auth in the host.
- **Never hold requests open** — always persist before returning from the start or resume path.
- **Retiring a blocked run** still requires the workflow version to be loaded on restore (versioning rules apply).
- Ship **design + bridge pattern** first; at most **one** new resume primitive if the bridge is not enough.
