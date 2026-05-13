# embed-kyc-service

A **credible backend slice** in one `go run`: human step → **persist** while blocked → **reload** from `pkg/run` as if a new request arrived → **`SubmitInput`** → run a **stub** “liveness” action → reach a terminal step → **save** again → print the **event trace**.

There is no HTTP server here. The point is the **library contract**: your app owns transport and storage; Maestro owns the graph, variables, and trace.

## Why it exists

**`library-basic`** shows decode + validate + first `RunUntilBlocked`. This demo answers the next questions people actually ask:

- Where does **`SubmitInput`** go after a human step?
- How do I **snapshot** an instance and **restore** it later?
- What does **`MemoryStore`** (optimistic **`revision`**) feel like in practice?

The workflow is embedded (`//go:embed workflow.yaml`) so the example is **self-contained** and still looks like something you would ship next to your service binary.

## Run

From the repository root:

```bash
go run ./examples/demos/embed-kyc-service
```

Output is **deterministic** for a fixed workflow: run id **`run_demo`**, scripted profile name, same graph every time. You should see **blocked at: collect-profile**, a **submitting profile input** stanza, **continued to: run-checks**, **completed at: approved**, then a **trace** block (`step.entered`, `run.blocked`, `input.accepted`, `action.ran`, `transition.taken`, `run.completed`, …).

## Takeaways for your app

This is a **narrow, scripted story** so you can see persistence and restore in one file—not a full product. Your API shape, tenancy, and database schema will differ. Patterns that usually **remain useful** when you rewrite it for real:

- **`run.NewMemoryStore`** is for tests and learning; in production implement **`run.Store`** (Postgres, Redis, your own table).
- Always reload with the **same** workflow definition metadata the run was created with (`id` / `version` and compatible graph).
- Serialize access per **`runID`** unless you know your store and engine usage are safe under concurrency.

## Files

| File | Role |
|------|------|
| `main.go` | Thin entrypoint. |
| `scenario.go` | The scripted “two requests” story + helpers. |
| `workflow.yaml` | Small KYC-shaped graph: profile → checks stub → **approved**. |
