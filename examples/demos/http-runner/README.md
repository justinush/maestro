# http-runner

Shows Maestro calling **real HTTP** from a workflow: **`RegistryWithHTTP`**, an **`http`** action with **`resultVariable`**, and a **CEL `when`** guard that reads **`variables.vendor.statusCode`**.

A tiny **`httptest.Server`** stands in for a vendor so the demo stays **offline** and **one command**.

## Why it exists

Stub actions are easy to read but they do not prove integrations. This folder is the bridge between “orchestration in YAML” and “my step hits an actual URL and branches on the response.”

## Run

From the repository root:

```bash
go run ./examples/demos/http-runner
```

`workflow.yaml` embeds a placeholder URL, **`__HTTP_BASE__`**. `main.go` substitutes the mock server’s URL before parse so we never hard-code a host in the workflow file.

**Execution:** **`RunUntilBlocked()`** returns a **`RunResult`**. The demo expects **`RunCompleted`** with **`res.StepID == "approved"`**; failures use **`RunFailed`** and **`res.Err`**.

You should see a **vendor snapshot** (status, headers, body), a short **trace** ending in **`run.completed`**, and a closing **`ok:`** line.

## Takeaways for your app

The mock server and **`%#v`** logging exist to keep the demo small. Your vendor URLs, auth, retries, and observability will not look like this file-for-file. Habits that **often carry over** when you build your version:

- Use **`httptest`** in unit tests; use a configured **`http.Client`** (timeouts, TLS, tracing) in production.
- **`srv.Client()`** is required with **`httptest`** so requests route to the test server correctly.
- Prefer **`json.MarshalIndent`** (or your own struct) over **`%#v`** when you log **`variables["vendor"]`**—`%#v` is fine for this demo but noisy in real logs.
- Use **`RunResult`** from **`RunUntilBlocked()`** to distinguish **done** vs **blocked** vs **failed** without sentinel **`error`** checks.

## Files

| File | Role |
|------|------|
| `main.go` | Wire mock, parse workflow, run engine, print trace. |
| `mock_vendor_server.go` | Returns **200** + JSON on **`/v1/liveness/status`**. |
| `workflow.yaml` | One action step, one terminal, guard on **`statusCode == 200`**. |
