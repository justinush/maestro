# Roadmap

Rough direction for Maestro. No dates — just what we're thinking about.

See the [changelog](../CHANGELOG.md) for what's already shipped.

## In 0.1.0 (done)

- Embed workflows in Go: human steps, actions, CEL guards
- `pkg/maestro` load / run / restore
- Persist runs (`MemoryStore`, Postgres adapter)
- Validate + simulate CLI
- Examples and a KYC backend demo

## In 0.2.0 (in progress)

- Workflow registry (`pkg/workflow`) — host multiple definitions; `LoadDir`, `Registry`, restore by `RunRecord` workflow id/version
- Workflow versioning — [design](design/workflow-versioning.md) (host policy for multiple versions)
- Async callbacks / webhooks — [design](design/async-callbacks.md) (bridge validated; custom action ergonomics next)

## Next (v0.2-ish)

- Custom action types in workflow YAML (app-owned `onEnter` runners; schema/validate ergonomics) — follow-up to [async callbacks](design/async-callbacks.md#next-core-focus-custom-actions) spike
- Retries and timers on action steps
- Better observability and typed variable helpers

## Maybe later

- More store adapters if people ask (Postgres is the reference for now)
- OpenTelemetry hooks
- Extra linting beyond JSON Schema

## Not on the table (for now)

Maestro stays a library you embed — not a product:

- Hosted orchestration SaaS
- Distributed worker cluster as a built-in thing
- Visual workflow editor

## Want to steer this?

Open a discussion or issue with your use case (KYC, onboarding, approvals, etc.). Real integration pain beats abstract feature requests.
