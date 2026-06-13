# Async callbacks design (v0.2.x)

How to resume a workflow days later — when a webhook fires, a manager clicks a link, or a payment provider calls back — without holding an HTTP request open.

This doc covers **waiting and resume**: what Maestro does today, what production hosts need, and what we may add to the engine next. Persistence and restore mechanics are in [architecture — pause / resume](../architecture.md#pause-resume-model). Workflow identity on restore follows [workflow versioning](workflow-versioning.md).

**Status:** design draft. Validate the bridge pattern in a real host integration before adding engine primitives such as `kind: wait`.

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
| Pause after outbound vendor call, resume on webhook | Action steps do not yield mid-step; HTTP runner is synchronous |
| Express “wait for external input” in the graph | Only `kind: human` blocks (resume step in prose) |
| Webhook payload → variables → transition | Works when the run is blocked on a resume step that accepts `SubmitInput` |

Vendor-shaped async is the main gap. UI- and API-driven resume on a blocking input step already works.

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

We may later add a dedicated step kind or signal API. Until then, the bridge models vendor waits as a resume step — blocking semantics already match.

---

## Callback identity

Maestro does not receive webhooks. The host does. Correlation is split on purpose.

### What Maestro stores

```text
runId              — primary key on RunRecord
currentStepId      — in Snapshot; resume step that expects the next payload
workflowId + version — on RunRecord for RestoreInstance
variables          — may include externalRef from an earlier create step (host convention)
```

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

1. Create step stores `externalRef` in `variables` (e.g. from HTTP action `resultVariable`) and registers `externalRef → runId` in a host table.
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
| **Graph (bridge)** | `action` step: `onEnter` HTTP POST creates session, stores `externalRef` in `variables` → resume step → transitions on webhook payload |
| **Wait** | Blocked on resume step (e.g. `wait-vendor-result`) |
| **Resume** | Webhook → `externalRef` lookup → restore → `SubmitInput` → `RunUntilBlocked` |
| **Engine change** | None for bridge |

**Create-step idempotency:** outbound `onEnter` actions that create vendor sessions must be idempotent at the host or vendor layer. Use idempotency keys such as `runId` + `stepId` so retries or partial failures do not open duplicate external resources. Store the returned `externalRef` in `variables` and in the host lookup table before blocking on the resume step.

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

**Outbound idempotency:** the create step’s `onEnter` may call an external API. If the host retries or re-enters that path, duplicate vendor sessions are a real risk. Create calls should be idempotent (e.g. idempotency key = `runId` + `create-verification` step id). Persist `externalRef` to `variables` and the host lookup table before advancing to the resume step.

**Pros:** ships on today’s engine; explicit create vs wait in the graph; `inputSchema` documents the webhook payload; `SubmitInput` stays the single resume API.

**Cons:** YAML still says `kind: human` for vendor waits; `presentationRef` is usually omitted on resume steps.

This is the **intended bridge** for v0.2.x — validate in a host integration before adding `kind: wait` or other engine primitives.

---

## Design options (engine — deferred)

**Do not implement `kind: wait` yet.** Validate the bridge in a host integration first (create session → resume step → webhook → `SubmitInput`). Revisit this table only if that spike proves the bridge is too awkward.

| Option | Idea | Pros | Cons |
|--------|------|------|------|
| **Bridge (default)** | Resume step + `SubmitInput` | Zero engine change; transport-neutral resume | YAML says `kind: human` |
| **`kind: wait`** | New step kind; blocks like resume step today | Clearer graph | Schema + engine loop change — **deferred** |
| **`DeliverSignal(name, payload)`** | Signal while blocked | Webhook-shaped naming | New API; overlap rules with `SubmitInput` |
| **Async action runner** | Action returns pending; new `RunStatus` | Single-step mental model | Harder to test; couples with retries |

**Current lean:** bridge only until host validation says otherwise. Avoid async-only action magic and avoid shipping multiple new primitives at once.

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
| **1 — design** | This doc + architecture / roadmap links | Agree on rules before code |
| **2 — bridge validation** | Host integration spike | Create vendor session → resume step → webhook → `SubmitInput`; **no engine change** |
| **3 — engine (only if spike fails)** | One resume primitive | `kind: wait` **or** `DeliverSignal` — not planned until bridge is proven insufficient |

Explicitly **not** in scope: inbound webhook server, vendor-specific adapters, callback DB inside Maestro, `kind: wait` before bridge validation.

---

## Open questions

1. **After bridge spike** — is `kind: human` resume step good enough long-term, or do we need `kind: wait` / `DeliverSignal`?
2. **`inputSchema` on resume steps** — required for vendor payloads, or optional with host validation only?
3. **Status subtyping** — does `RunBlocked` need a reason (`ui` vs `callback`), or is step id + host context enough?
4. **Simulate tooling** — `maestro simulate` inject resume step via existing submit vs new command?

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
- **Vendor async (bridge):** action create (idempotent `onEnter`) + resume step + webhook → `externalRef` lookup → `SubmitInput`.
- **Callback identity:** `runId` + `stepId` in Maestro; `externalRef` → `(runId, expectedStepId)` in the host.
- **Never hold requests open** — always persist before returning.
- **Defer `kind: wait`** — validate bridge in a host spike first; engine primitives only if that fails.
