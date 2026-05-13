# library-basic

Smallest possible Maestro embedding example.

## What it demonstrates

```txt
Decode workflow → Validate → NewInstance → RunUntilBlocked → (completed | ErrNeedsInput)
```

The program drives the instance in a loop with **`RunUntilBlocked`** until the workflow either **finishes** or **pauses for a human**—not a single fixed number of calls.

## Run

From the repository root:

```bash
go run ./examples/demos/library-basic examples/workflows/workflow-v0-minimal.yaml
```

## Expected output

With the default **`workflow-v0-minimal.yaml`**, the first blocking step is **`collect-profile`**. You should see exactly:

```txt
blocked at "collect-profile" (needs input — see examples/demos/embed-kyc-service)
```

If you point at a different workflow, the step id or outcome may differ.

## Why it stops here

This demo shows the smallest embedding path. It intentionally stops at the first **`human`** step so you can see how Maestro **pauses** the run and returns **`ErrNeedsInput`**, handing control back to your application (your API, worker, or UI would call **`SubmitInput`** next).

## What this demo does not show

- Submitting human input (**`SubmitInput`**)
- Persisting snapshots or **`pkg/run`** stores
- Restoring a run from storage
- Custom **`ActionRegistry`** runners (beyond default **`stub`**)
- HTTP or other external integrations

For persistence + **`SubmitInput`**, see **`../embed-kyc-service`**. For HTTP actions, see **`../http-runner`**.

## Takeaways for your app

Treat this folder as a **learning sketch**, not code to paste unchanged into production—your routes, storage, and workflow IDs will look different. The ideas that usually **still help** when you adapt it:

- **`pkg/definition`** — read the workflow YAML/JSON into a **`WorkflowDefinition`**.
- **`pkg/validate`** — run the same checks as **`maestro validate`** so bad files fail before you create an instance.
- **`pkg/engine`** — create an **`Instance`**, then drive it with **`RunUntilBlocked`**. When a step needs a person, call **`SubmitInput`** after **`ErrNeedsInput`** (see **`../embed-kyc-service`**).

**`DefaultRegistry()`** here only knows built-in **`stub`** actions. When you add real side effects (HTTP calls, queues, your own integrations), build a **`Registry`**, **`Register`** your action types, and pass that registry in **`engine.Options`** instead of the default.
