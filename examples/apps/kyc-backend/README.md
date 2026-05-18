# kyc-backend

Minimal HTTP API that embeds Maestro like a real KYC service.

This demo wires REST handlers to `pkg/maestro` and `pkg/run`. Maestro owns the workflow graph; your app owns applicants, documents, and UI-facing status.

```txt
POST /kyc/start
GET  /kyc/{runID}
GET  /kyc/{runID}/events
POST /kyc/{runID}/profile
POST /kyc/{runID}/document
POST /kyc/{runID}/review
```

Mutating handlers follow: restore run → validate step → `SubmitInput` → `RunUntilBlocked` → save workflow → save app data → return status.

```txt
collect-profile -> document-upload -> run-liveness-check -> approved
                                              \-> manual-review -> approved
```

Passport uploads set `review.required` and route to manual review; other document types can auto-approve after liveness.

---

## What this demo shows

- embedding Maestro behind HTTP handlers
- persisting and restoring workflow runs per request
- `SubmitInput` on human steps from API payloads
- app-owned data saved only after workflow success
- mapping `currentStepId` to UI-friendly `status`
- execution trace via `GET /kyc/{runID}/events`

Use this after the three demos under `examples/demos/` when you need a request/response shape, not just a scripted `main`.

---

## Run

From the repository root:

```bash
go run ./examples/apps/kyc-backend
```

Server listens on `:8080` (override with `ADDR`).

---

## Expected flow (curl)

```bash
curl -s -X POST http://localhost:8080/kyc/start | jq
export RUN_ID=run_xxxxxxxx

curl -s -X POST "http://localhost:8080/kyc/${RUN_ID}/profile" \
  -H 'Content-Type: application/json' \
  -d '{"fullName":"Demo User","email":"demo@example.com"}' | jq

curl -s -X POST "http://localhost:8080/kyc/${RUN_ID}/document" \
  -H 'Content-Type: application/json' \
  -d '{"documentType":"national_id","documentRef":"doc-123"}' | jq

curl -s "http://localhost:8080/kyc/${RUN_ID}" | jq
curl -s "http://localhost:8080/kyc/${RUN_ID}/events" | jq
```

Passport path (manual review):

```bash
curl -s -X POST "http://localhost:8080/kyc/${RUN_ID}/document" \
  -H 'Content-Type: application/json' \
  -d '{"documentType":"passport","documentRef":"pass-456"}' | jq

curl -s -X POST "http://localhost:8080/kyc/${RUN_ID}/review" \
  -H 'Content-Type: application/json' \
  -d '{"approved":true}' | jq
```

Typical statuses: `awaiting_profile`, `awaiting_document`, `awaiting_review`, `approved`.

---

## Files

| File | Purpose |
|---|---|
| `main.go` | HTTP server entrypoint |
| `handler.go` | REST routes + strict JSON decode |
| `service.go` | workflow-first submit lifecycle |
| `errors.go` | sentinel errors for HTTP mapping |
| `status.go` | step → UI `status` |
| `applicant_store.go` | app-owned applicant data |
| `maestro_store.go` | run persistence helpers |
| `vendor.go` | fake vendor hook (by applicant id) |
| `workflow.go` | embedded workflow loader |
| `workflow.yaml` | sample KYC workflow |

---

## Real-world mapping

| Piece | Role |
|---|---|
| HTTP handlers | your API layer |
| `ApplicantStore` | your DB tables / CRM |
| `run.Store` | workflow snapshot storage |
| `SubmitInput` | form submission per screen |
| `status.go` | response DTOs for the UI |
| `/events` | support / debug trace |

In production, replace `MemoryStore` with a real database and swap the liveness stub for an HTTP action (see `examples/demos/http-runner`).

---

## Related demos

- [`library-basic`](../../demos/library-basic) — smallest embed path
- [`embed-kyc-service`](../../demos/embed-kyc-service) — persist + restore without HTTP
- [`http-runner`](../../demos/http-runner) — vendor HTTP from the workflow
