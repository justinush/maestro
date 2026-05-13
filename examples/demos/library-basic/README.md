# library-basic

The smallest useful Maestro program: **load a workflow file**, **validate** it the same way the CLI does, spin up an **engine instance**, and call **`RunUntilBlocked` once**.

## Why it exists

Most services eventually hit a **human** step (a form, a review queue, a mobile capture flow). At that point the engine stops with **`ErrNeedsInput`**. This demo stops there on purpose so you can see the skeleton without scrolling through persistence code.

If you want the full loop—**`SubmitInput`**, **`RunUntilBlocked`**, and **`pkg/run`**—open **`../embed-kyc-service`**.

## Run

From the repository root:

```bash
go run ./examples/demos/library-basic examples/workflows/workflow-v0-minimal.yaml
```

You should see a line ending with **blocked at `"collect-profile"`** and a pointer to **`embed-kyc-service`**.

## Takeaways for your app

Treat this folder as a **learning sketch**, not code to paste unchanged into production—your routes, storage, and workflow IDs will look different. The ideas that usually **still help** when you adapt it:

- **`pkg/definition`** — read the workflow YAML/JSON into a **`WorkflowDefinition`**.
- **`pkg/validate`** — run the same checks as **`maestro validate`** so bad files fail before you create an instance.
- **`pkg/engine`** — create an **`Instance`**, then drive it with **`RunUntilBlocked`**. When a step needs a person, call **`SubmitInput`** after **`ErrNeedsInput`** (see **`../embed-kyc-service`**).

**`DefaultRegistry()`** here only knows built-in **`stub`** actions. When you add real side effects (HTTP calls, queues, your own integrations), build a **`Registry`**, **`Register`** your action types, and pass that registry in **`engine.Options`** instead of the default.
