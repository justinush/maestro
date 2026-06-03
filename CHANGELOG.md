# Changelog

Notable Maestro changes live here.

## [0.2.0] - TBD

### Added

- `pkg/workflow`: workflow registry for hosting multiple workflow definitions in one process (`LoadDir`, `Registry`, lookup by `id` + `version`, `RestoreInstance` with identity checks). Single-workflow `maestro.Load` / `Runtime` APIs are unchanged.

## [0.1.1] - 2026-05-31

### Fixed

- `maestro --version` shows the real tag when you install with `go install ...@vX.Y.Z` (was always `dev` in v0.1.0)

### Added

- Cross-platform CLI archives on GitHub Releases (linux, darwin, windows; amd64 and arm64 where applicable)
- `SHA256SUMS` alongside release archives
- `make verify-version` and `make build-release` for maintainers

### Notes

`go install` and release binaries should both print the module tag (e.g. `v0.1.1`).

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

Next up: a clearer versioning story, then async callbacks, retries/timers, and observability (`pkg/workflow` registry ships in v0.2.0).
