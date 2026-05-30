# Roadmap

Rough direction for Maestro. No dates — just what we're thinking about.

See the [changelog](../CHANGELOG.md) for what's already shipped.

## In 0.1.0 (done)

- Embed workflows in Go: human steps, actions, CEL guards
- `pkg/maestro` load / run / restore
- Persist runs (`MemoryStore`, Postgres adapter)
- Validate + simulate CLI
- Examples and a KYC backend demo

## Next (v0.2-ish)

- Workflow registries (host multiple workflow definitions in one application)
- Clearer workflow versioning story (new YAML vs runs already in flight)
- Async callbacks / webhooks (resume a workflow without holding a request open)
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
