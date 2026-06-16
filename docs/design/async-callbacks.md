# Async callbacks design (v0.2.x)

How to resume a workflow days later — when a webhook fires, a manager clicks a link, or a payment provider calls back — without holding an HTTP request open.

This doc covers **waiting and resume**: what Maestro does today, what production hosts need, and what we may add to the engine next. Persistence and restore mechanics are in [architecture — pause / resume](../architecture.md#pause-resume-model). Workflow identity on restore follows [workflow versioning](workflow-versioning.md).

**Status:** bridge pattern **validated** in a host integration spike (phase 2 complete). No new engine primitive for async resume yet. Next core focus: [custom action extensibility](#next-core-focus-custom-actions) so outbound create steps model cleanly in YAML.

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
blocking input step → RunUntilBlocked → RunBlocked
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

**Delayed resume** is already async. A **blocking input step** waits until the host delivers a payload. In today’s schema that is `kind: human`; the same mechanism covers UI forms, webhooks, approval links, and internal APIs.

When execution reaches such a step, `RunUntilBlocked` returns `RunBlocked`. The host persists `RunRecord` and returns. Hours or weeks later:

```go
rec, _ := store.Get(ctx, runID)
in, _ := rt.RestoreInstance(rec, maestro.InstanceOptions{ActionRegistry: reg})

sub := in.SubmitInput(payload)          // validate, merge variables, advance if a transition fires
_ = in.RunUntilBlocked()                // drive action steps until next block or terminal
_ = store.Save(ctx, run.RecordFromInstance(in, def, rec.Revision))
```

**`SubmitInput` is transport-neutral.** The name reflects step kind in YAML, not where the payload comes from. A host may call it from a UI handler, webhook endpoint, signed approval link, or backoffice API — same validation, same transitions.

The [kyc-backend demo](../../examples/apps/kyc-backend/) follows this pattern: `Start` blocks on the first blocking input step; later API calls restore and `SubmitInput` on the expected step.

**Correlation id:** `RunID` on the instance and in `RunRecord` / `Snapshot`. That is the primary key hosts use to tie a callback to a run.

**Concurrency:** `RunRecord.Revision` + `Store.Save` optimistic locking — two concurrent resumes must not silently clobber each other.

So: **approval-link tomorrow** and **form submit next week** are not missing engine features. They need host HTTP handlers and persistence discipline, not a new primitive. Vendor webhooks use the same `SubmitInput` path once the run is on a **resume step** (see [Bridge pattern](#bridge-pattern) below).

---

## The gap

| Need | Today |
|------|--------|
| Pause after outbound vendor call, resume on webhook | **Bridge:** separate action create step + resume step; no mid-action pause. Create must be idempotent (`runId` + `stepId`). |
| Express “wait for external input” in the graph | `kind: human` blocks (resume step). Webhook-only waits use placeholder `presentationRef` until/unless `kind: wait` is revisited. |
| Webhook payload → variables → transition | Works when blocked on resume step with `SubmitInput` — **validated in host spike**. |
| Model custom outbound create in YAML | **Gap:** v0.1 schema allows only `stub` / `http`; engine supports custom runners. Custom action ergonomics is next. |

Vendor-shaped async resume is **solved by the bridge**. Remaining friction is **schema ergonomics** for create steps, not a missing resume primitive.

---

## Waiting state (engine model)

A run is **waiting** when the host must not call `RunUntilBlocked` again until it delivers external input.

| Field | Meaning |
|-------|---------|
| `RunResult.Status == RunBlocked` | Engine yielded on a blocking input step (`kind: human` today) |
| `Snapshot.CurrentStepID` | Resume step — must receive the next payload |
| `Snapshot.OnEnterRan` | Whether `onEnter` actions on that step already ran |
| `Snapshot.Variables` | State merged from prior steps and prior inputs |
| `RunRecord.WorkflowID` / `WorkflowVersion` | Definition used on restore (exact version — see versioning doc) |

Trace event `run.blocked` (`EventBlocked`) is recorded when blocking.

**Contract:** every host request that ends in a wait must **persist** the snapshot before returning. Never keep a process open until a webhook arrives.

### Waiting kinds (conceptual)

| Kind | Who delivers resume | Engine API today |
|------|---------------------|------------------|
| **UI / API input** | End user, backoffice API | `SubmitInput` on resume step |
| **Async callback** | Webhook handler, approval link, provider callback | `SubmitInput` on resume step (bridge) |

Both paths use the same engine API. The host chooses the transport.

We may later add a dedicated step kind or signal API if the bridge proves insufficient in practice. The host spike showed `kind: human` resume steps are enough for webhook resume; placeholder `presentationRef` is acceptable for now.

---

## Callback identity

Maestro does not receive webhooks. The host does. Correlation is split on purpose.

### What Maestro stores

```text
runId              — primary key on RunRecord
currentStepId      — in Snapshot; resume step that expects the next payload
workflowId + version — on RunRecord for RestoreInstance
variables          — workflow state; may optionally mirror externalRef (see below)
```

### `externalRef` mapping (host contract)

The host owns callback routing. Maestro does not store vendor session ids, payment ids, or webhook tokens.

**Canonical mapping (recommended):** a host database table is the source of truth:

```text
externalRef  →  (runId, expectedStepId)
```

Use this table to resolve inbound webhooks and provider callbacks. Index by `externalRef` for lookup; optionally index by `runId` for status APIs.

**Optional mirror in `variables`:** after create, the host may also write `externalRef` into workflow variables (e.g. from an HTTP action `resultVariable` or a custom action runner). Useful for UI, debug, and tracing — but **do not** treat `variables` as the only mapping. Hosts that only store `externalRef` in variables lose a stable lookup key when enriching API responses from a blocked snapshot.

**Idempotency:** create calls should use keys such as `runId` + `stepId` so retries do not open duplicate vendor resources. Webhook redelivery should be deduplicated on vendor `eventId` (or equivalent) in the host before calling `SubmitInput` again.

### What the host adds

```text
externalRef lookup table:
  vendorSessionId / providerPaymentId / …  →  (runId, expectedStepId)

Signed token in callback URL  (runId + stepId + expiry + HMAC)
Webhook signature verification
Idempotency on vendor event id (safe to process duplicate delivery)
```

**Minimum resume key (Maestro):** `(runId, stepId)` — after restore, the host checks `in.CurrentStepID()` matches the step the callback is for. Wrong step → reject (see kyc-backend `ErrWrongStep` pattern).

**Typical vendor flow:**

1. Create step (action `onEnter`) creates the vendor session idempotently and registers `externalRef → (runId, expectedStepId)` in the host table. Optionally mirror `externalRef` into `variables`.
2. Webhook body carries the vendor’s id → host lookup → `runId` + `expectedStepId`.
3. Restore → step guard → `SubmitInput(normalizedPayload)`.

Maestro does not plan a global callback registry in core. The host owns the `externalRef` mapping table and URL shapes.

---

## Host resume recipe

Canonical handler for any async completion (approval, webhook, provider callback):

```go
// 1. Resolve run (from URL token, or externalRef lookup from webhook body)
rec, err := store.Get(ctx, runID)

// 2. Restore exact workflow version
in, err := reg.RestoreInstance(rec, maestro.InstanceOptions{ActionRegistry: reg})

// 3. Guard: callback is for the resume step we expect
if in.CurrentStepID() != wantStep {
    return errWrongStep
}

// 4. Deliver payload — SubmitInput is transport-neutral (UI, webhook, link, API)
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

**Host owns:** steps 1 and 3 (auth, `externalRef` mapping, idempotency, normalizing vendor JSON to the input map).

**Maestro owns:** steps 4–6 semantics (validation against `inputSchema`, transitions, action execution, snapshot).

**Callback idempotency:** if the vendor redelivers the same event, the host should detect “already advanced past this step” or “duplicate event id” before calling `SubmitInput` again. Engine-level dedup is not in scope for the first slice.

---

## Maestro vs host

| Maestro | Host |
|---------|------|
| `RunBlocked` / `RunCompleted` / `RunFailed` | HTTP routes for webhooks and approval links |
| Snapshot + `RunRecord` shape | Vendor SDKs and outbound API calls |
| `SubmitInput` on blocking input steps | Vendor payload → input map (webhook, UI, link, API) |
| `inputSchema` validation on resume | Webhook auth, replay protection |
| Execution trace (`run.blocked`, `input.accepted`, …) | When to call vendor (which step’s `onEnter`) |
| Restore by exact `(workflowId, version)` | Retry polling, DLQ, alerting |

---

## Scenario patterns

### Manager approval (link clicked tomorrow)

| | |
|--|--|
| **Graph** | Resume step (`kind: human`) with `inputSchema` for `{ approved: bool }` |
| **Wait** | `RunUntilBlocked` → `RunBlocked` → persist |
| **Resume** | Link hits host → restore → `SubmitInput({ approved: true })` → `RunUntilBlocked` |
| **Engine change** | None required |

`SubmitInput` from the link handler — same API as a backoffice form post.

### Vendor verification (create session, webhook days later)

| | |
|--|--|
| **Graph (bridge)** | `action` step: `onEnter` creates session (HTTP or app-owned action type) → resume step → transitions on webhook payload |
| **Wait** | Blocked on resume step (e.g. `wait-vendor-result`) |
| **Resume** | Webhook → `externalRef` lookup → restore → `SubmitInput` → `RunUntilBlocked` |
| **Engine change** | None for bridge (validated) |

**Create-step idempotency:** outbound `onEnter` actions that create vendor sessions must be idempotent at the host or vendor layer. Use idempotency keys such as `runId` + `stepId` so retries or partial failures do not open duplicate external resources. Register `externalRef` in the host lookup table (and optionally in `variables`) before blocking on the resume step.

### Payment provider callback

Same as vendor verification: outbound create + resume step + webhook → `SubmitInput`. Provider signature verification, `externalRef` mapping, and idempotency stay in the host.

---

## Bridge pattern

Until we add a dedicated step kind or signal API, model vendor waits as two steps: **create** (action) + **resume** (blocking input step that waits for external input).

In YAML the resume step is still `kind: human`. In prose we call it a **resume step** — it accepts input from a webhook as readily as from a UI.

```yaml
- id: create-verification
  kind: action
  onEnter:
    - type: http
      id: create-session
      params:
        method: POST
        url: "https://vendor.example/v1/sessions"
        resultVariable: vendor   # host reads vendor.sessionId → externalRef

- id: wait-vendor-result
  kind: human
  presentationRef: internal/wait-vendor@v1   # placeholder when resume is webhook-only (see below)
  description: Resume step — blocks until vendor webhook delivers result.
  inputSchema:
    type: object
    required: [status]
    properties:
      status:
        type: string
      externalRef:
        type: string

transitions:
  - from: create-verification
    to: wait-vendor-result
  - from: wait-vendor-result
    to: next-step
    when: "variables.status == 'approved'"
```

**Outbound idempotency:** the create step’s `onEnter` may call an external API. If the host retries or re-enters that path, duplicate vendor sessions are a real risk. Create calls should be idempotent (e.g. idempotency key = `runId` + `create-verification` step id). Register `externalRef` in the host lookup table (and optionally in `variables`) before advancing to the resume step.

**`presentationRef` on webhook-only resume steps:** v0.1 schema requires `presentationRef` on `kind: human`. When the only consumer of the resume step is a webhook (no UI), use a host-internal placeholder such as `internal/wait-vendor@v1`. The host does not need to render it. If many integrations need this pattern, revisit a dedicated step kind (e.g. `kind: wait`) later — not required for the bridge.

**Pros:** ships on today’s engine; explicit create vs wait in the graph; `inputSchema` documents the webhook payload; `SubmitInput` stays the single resume API.

**Cons:** YAML still says `kind: human` for vendor waits; `presentationRef` is a placeholder when there is no UI.

This is the **intended bridge** for v0.2.x. Host integration spike **validated** it — see [Bridge validation](#bridge-validation-spike-outcome).

---

## Bridge validation (spike outcome)

A separate-module host integration exercised the full path with **no Maestro engine changes**:

```text
POST start  →  action create step  →  resume step (RunBlocked)  →  persist
POST webhook  →  externalRef lookup  →  restore  →  SubmitInput  →  RunUntilBlocked  →  save
```

**Validated:**

- `SubmitInput` is sufficient for webhook resume (transport-neutral).
- `kind: human` resume step blocking semantics match multi-day vendor waits.
- Host-owned `externalRef` lookup table + webhook `eventId` deduplication work as designed.
- `workflow.Registry` + `run.Store` restore after persist behaves correctly on the happy path.

**Awkward (schema, not engine):** the create step could not be modeled as a custom action type (e.g. `type: vendor-create-session`) because v0.1 JSON Schema only allows `stub` and `http` in YAML. The engine already supports app-owned action runners via `engine.Registry.MustRegister`; validation is the gap. The spike used a `stub` in YAML and ran idempotent create in host code after `RunUntilBlocked` — workable, but create is invisible in the graph.

**Verdict:** bridge pattern is **good enough**. Do **not** add `kind: wait`, `DeliverSignal`, async action runners, or an engine callback registry yet.

**Next core focus:** [custom action extensibility](#next-core-focus-custom-actions) so outbound create steps bind to app-owned runners in YAML. That removes the biggest spike friction without a new resume primitive.

---

## Next core focus: custom actions

Runtime already dispatches custom action types when the host registers them (`engine.Registry.MustRegister`). v0.1 schema validation does not — `onEnter[].type` is enum-locked to `stub` and `http`.

Planned follow-up (docs + schema/validate ergonomics, not a new execution primitive):

1. Allow app-defined action types in workflow YAML (e.g. `type: vendor-create-session`) when the host registers the runner and opts in at `workflow.LoadDir` / validate time.
2. Keep `stub` and `http` strictly validated; custom types pass structural checks and fail at run time if unregistered (same as today’s engine behavior).
3. Document the embedding recipe: register runner → allow type in validator → reference in YAML.

Once that lands, the bridge graph can express create in YAML again instead of host-side workarounds after block.

**Explicitly deferred** (unchanged after spike): `kind: wait`, `DeliverSignal`, async action runner, engine callback registry.

---

## Design options (engine — deferred)

**Bridge validated.** Do not implement the alternatives below unless a future integration proves the bridge insufficient (e.g. widespread pain from placeholder `presentationRef` alone).

| Option | Idea | Pros | Cons | Status |
|--------|------|------|------|--------|
| **Bridge (default)** | Resume step + `SubmitInput` | Zero engine change; transport-neutral resume | YAML says `kind: human`; placeholder `presentationRef` for webhooks | **Shipped pattern** |
| **`kind: wait`** | New step kind; blocks like resume step today | Clearer graph for webhook-only waits | Schema + engine loop change | **Deferred** |
| **`DeliverSignal(name, payload)`** | Signal while blocked | Webhook-shaped naming | New API; overlap rules with `SubmitInput` | **Deferred** |
| **Async action runner** | Action returns pending; new `RunStatus` | Single-step mental model | Harder to test; couples with retries | **Deferred** |
| **Engine callback registry** | Core stores `externalRef` → run | Central lookup | Wrong layer; host owns vendors | **Not planned** |

**Current lean:** bridge + host mapping table. Improve YAML ergonomics via custom action types before revisiting new resume primitives.

Relationship to **retries / timers** (roadmap): callbacks are **push** (event arrives). Retries are **pull** (engine or host tries again). They compose but ship as separate designs.

---

## Relationship to shipped APIs

Already available:

- `Instance.RunUntilBlocked()` → `RunBlocked` on blocking input steps (`kind: human` today)
- `Instance.SubmitInput(map[string]any)` — transport-neutral resume on those steps
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
| **1 — design** | This doc + architecture / roadmap links | Done |
| **2 — bridge validation** | Host integration spike | **Done** — create → resume step → webhook → `SubmitInput`; no engine change |
| **3 — custom actions** | Schema / validate ergonomics for app-owned `onEnter` types | Next — see [Next core focus](#next-core-focus-custom-actions) |
| **4 — engine (only if bridge fails)** | `kind: wait` **or** `DeliverSignal` | **Not planned** after spike; revisit only if bridge proves insufficient |

Explicitly **not** in scope: inbound webhook server, vendor-specific adapters, callback DB inside Maestro, `kind: wait` before custom-action ergonomics land.

---

## Open questions

1. **After bridge spike** — **resolved:** `kind: human` resume step is good enough for async callbacks. Defer `kind: wait` / `DeliverSignal` unless placeholder `presentationRef` becomes painful across many hosts.
2. **`inputSchema` on resume steps** — spike used required `status` on vendor resume step; keep as host best practice for webhook payloads. Still optional in schema.
3. **Status subtyping** — open: does `RunBlocked` need a reason (`ui` vs `callback`), or is step id + host context enough? Spike used step id + host status mapping only.
4. **Simulate tooling** — open: `maestro simulate` inject resume step via existing submit vs new command?
5. **Custom action types in YAML** — open: allowlist at `workflow.LoadDir` vs schema version bump; see [Next core focus](#next-core-focus-custom-actions).

---

## Related docs

- [Architecture — pause / resume](../architecture.md#pause-resume-model)
- [Architecture — persistence](../architecture.md#persistence-model)
- [Workflow versioning](workflow-versioning.md) — exact version on restore; long-running runs
- [Roadmap](../roadmap.md)
- [kyc-backend demo](../../examples/apps/kyc-backend/) — restore + `SubmitInput` reference host

---

## Summary

- **Blocking input / resume steps already work:** `RunBlocked` → persist → later `SubmitInput` → `RunUntilBlocked` → save.
- **`SubmitInput` is transport-neutral** — UI, webhook, approval link, or internal API.
- **Vendor async (bridge):** action create (idempotent `onEnter`) + resume step + webhook → `externalRef` lookup → `SubmitInput`. **Validated in a host spike.**
- **Callback identity:** `runId` + `stepId` in Maestro; host table `externalRef` → `(runId, expectedStepId)` as canonical mapping (optional `variables` mirror).
- **Never hold requests open** — always persist before returning.
- **Defer new resume primitives** — no `kind: wait` / `DeliverSignal` / async runner / engine callback registry yet.
- **Next:** custom action type extensibility in schema/validate so create steps (e.g. `vendor-create-session`) model cleanly in YAML.
