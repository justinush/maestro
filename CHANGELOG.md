# Changelog

Notable Maestro changes live here.

## [0.1.0] - 2026-05-30

First public release.

### Added

- Embed a workflow runtime in your Go backend — no separate orchestration server
- Human steps (pause for input) and action steps (stub, HTTP)
- Workflow YAML/JSON with validation and CEL guards on transitions
- `pkg/maestro` — load a workflow, run it, restore it on the next request
- Persist runs with `run.Store`, optimistic locking on `revision`
- `run.MemoryStore` for tests and demos
- `pkg/run/postgres` for real Postgres persistence (JSONB + optional `ApplySchema`)
- `maestro validate` and `maestro simulate` CLI
- Examples from minimal embed up to a KYC backend demo
- Architecture notes and contributing guide

### Notes

This release is the basics: run a workflow, hit a human step, save it, pick it up again later.

While we're still in v0.x, we want `RunRecord`, `Snapshot`, and run status values to stay predictable. Things can still change as we learn, but we won't casually break persistence shapes — that'll be a major version thing.

Next up: workflow registries, a clearer versioning story, then async callbacks, retries/timers, and observability.
