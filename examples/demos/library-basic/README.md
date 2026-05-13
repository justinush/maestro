# library-basic

Smallest possible Maestro embedding example.

## What it demonstrates

```txt
maestro.Load(path) → Runtime.NewInstance → RunUntilBlockedResult → (RunCompleted | RunBlocked)
```

This uses **`pkg/maestro`**: **`Load`** performs **decode + validate** in one step, then **`NewInstance`** starts the run. The loop calls **`RunUntilBlockedResult()`** and branches on **`res.Status`** (for example **completed** vs **waiting for input**).

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

This demo shows the smallest embedding path. It intentionally stops on the first **`human`** step: Maestro **pauses** the run and hands control back to your application (your API, worker, or UI would call **`SubmitInput`** next). In code, that pause shows up as **`engine.RunBlocked`** from **`RunUntilBlockedResult()`**.

## What this demo does not show

- Submitting human input (**`SubmitInput`**)
- Persisting snapshots or **`pkg/run`** stores
- Restoring a run from storage
- Custom **`ActionRegistry`** runners (beyond default **`stub`**)
- HTTP or other external integrations

For persistence + **`SubmitInput`**, see **`../embed-kyc-service`**. For HTTP actions, see **`../http-runner`**.

## Takeaways for your app

Treat this folder as a **learning sketch**, not code to paste unchanged into production—your routes, storage, and workflow IDs will look different. The ideas that usually **still help** when you adapt it:

- **`pkg/maestro`** — **`Load`** / **`LoadWithValidate`** / **`Compile`** for a shorter first path; **`Runtime.NewInstance`** with **`InstanceOptions`** when you need run ids, variables, or a custom registry.
- **`pkg/definition`**, **`pkg/validate`**, **`pkg/engine`** — use when you need full control (custom decode, validate flags, or **`NewInstance`** / **`RunUntilBlocked`** directly).
- **`RunUntilBlockedResult`** — each return includes **`Status`**, **`StepID`**, and **`Events`** for the trace so far; check **`res.Err`** when **`Status`** is **`RunFailed`**.

**`DefaultRegistry()`** (via empty **`InstanceOptions`**) only knows built-in **`stub`** actions. For HTTP or your own integrations, set **`InstanceOptions.ActionRegistry`** or call **`engine.NewInstance`** with a registry you built.
