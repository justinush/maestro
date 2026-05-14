# library-basic

Smallest possible Maestro embedding example.

## What it demonstrates

```txt
maestro.Load(path) → Runtime.NewInstance → RunUntilBlocked() → (RunCompleted | RunBlocked | RunFailed)
```

This uses **`pkg/maestro`**: **`Load`** performs **decode + validate** in one step, then **`NewInstance`** starts the run. A **single** **`RunUntilBlocked()`** call drives the engine until it stops; the return value is a **`RunResult`**—branch on **`res.Status`** (**`engine.RunCompleted`**, **`engine.RunBlocked`**, **`engine.RunFailed`**) and use **`res.StepID`** / **`res.Events`** as needed.

## Run

From the repository root:

```bash
go run ./examples/demos/library-basic examples/workflows/workflow-v0-minimal.yaml
```

## Expected output

With the default **`workflow-v0-minimal.yaml`**, you should see a **loaded** line (workflow **`id`** and **`version`** from the file), then the blocked line:

```txt
loaded workflow "example-onboarding" (version "1.0.0")
blocked at "collect-profile" (needs input — see examples/demos/embed-kyc-service)
```

If you point at a different workflow, the ids, version, step id, or outcome may differ.

## Why it stops here

This demo shows the smallest embedding path. It intentionally stops on the first **`human`** step: Maestro **pauses** the run and hands control back to your application (your API, worker, or UI would call **`SubmitInput`** next). In code, that pause is **`engine.RunBlocked`** on the **`RunResult`** from **`RunUntilBlocked()`**.

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
- **`pkg/definition`**, **`pkg/validate`**, **`pkg/engine`** — use when you need full control (custom decode, validate flags, or **`engine.NewInstance`** / **`RunUntilBlocked`** directly).
- **`RunUntilBlocked()`** — returns **`RunResult`**: **`Status`**, **`StepID`**, **`Events`** (trace snapshot); **`Err`** is set only when **`Status`** is **`RunFailed`**.

**`DefaultRegistry()`** (via empty **`InstanceOptions`**) only knows built-in **`stub`** actions. For HTTP or your own integrations, set **`InstanceOptions.ActionRegistry`** or call **`engine.NewInstance`** with a registry you built.
