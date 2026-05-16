# http-runner

Minimal example of calling external services from a Maestro workflow.

<p align="center">
  <img src="./.docs/assets/http-runner.excalidraw.png" alt="library-basic flow" width="960">
</p>

This demo shows a workflow action step calling a mock vendor API through Maestro's HTTP runner.

```txt
workflow action -> HTTP runner -> vendor API -> guard evaluation -> approved
```

The workflow stores the HTTP response in workflow variables, then evaluates a guard to decide the next step.

---

## What this demo shows

- HTTP-based workflow actions
- external service orchestration
- storing HTTP responses in workflow variables
- guard evaluation from vendor responses
- workflow-driven decision making

This is the foundation for integrating:

- KYC vendors
- AML services
- fraud/risk systems
- internal verification APIs
- compliance backoffice services

---

## Run

From the repository root:

```bash
go run ./examples/demos/http-runner
```

---

## Expected output

```txt
mock vendor started

loaded workflow "http-runner-demo" (version "1.0.0")

running workflow with HTTP runner...

completed at: approved

vendor HTTP status: 200
vendor liveness status: pass

trace:
...
```

Output may differ slightly as the engine evolves.

---

## Files

| File | Purpose |
|---|---|
| `main.go` | workflow execution scenario |
| `workflow.go` | embedded workflow loader |
| `workflow.yaml` | sample HTTP-based workflow |
| `mock_vendor_server.go` | local mock vendor API |

---

## Real-world mapping

This demo roughly maps to:

| Workflow step | Real-world equivalent |
|---|---|
| HTTP runner | call vendor/internal service |
| Mock vendor API | Onfido / Sumsub / AML API |
| `variables.vendor` | captured vendor response |
| Guard evaluation | workflow routing decision |

In production, the HTTP runner would typically call real vendor or internal APIs instead of the local `httptest.Server`.

---

## Related demos

- [`library-basic`](../library-basic) — smallest embedding example
- [`embed-kyc-service`](../embed-kyc-service) — persistence + restore lifecycle
