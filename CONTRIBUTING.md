# Contributing

Thanks for your interest in contributing to Maestro.

Maestro is still in early `v0.x` development. The current focus is keeping the runtime simple, understandable, and easy to embed into backend applications.

Before opening a large feature PR, please open an issue or discussion first.

---

# Development Setup

Requirements:

- Go 1.24+
- Git

Clone the repository:

```bash
git clone https://github.com/justinush/maestro.git
cd maestro
```

Run tests:

```bash
go test ./...
```

Run the examples:

```bash
go run ./examples/demos/library-basic examples/workflows/workflow-v0-minimal.yaml
```

```bash
go run ./examples/demos/embed-kyc-service
```

```bash
go run ./examples/demos/http-runner
```

```bash
go run ./examples/apps/kyc-backend
```

---

# Project Philosophy

Maestro prioritizes:

- embedded-first orchestration
- readable workflows
- human-in-the-loop flows
- backend-oriented integration
- simple runtime APIs
- understandable execution flow

In general:

- prefer clarity over abstraction
- prefer explicit lifecycle flow
- avoid hidden magic
- keep orchestration easy to reason about

---

# Pull Requests

Before opening a PR:

- run `go test ./...`
- keep examples working
- add tests when changing engine behavior
- avoid unrelated refactors in the same PR

Small focused PRs are much easier to review.

---

# Examples

The examples under `examples/` are part of the developer experience.

If you change:
- runtime APIs
- persistence behavior
- orchestration lifecycle
- execution semantics

please verify the examples still:
- compile
- run
- remain easy to understand

The examples are intentionally designed as progressive learning steps.

---

# API Design Notes

Before adding a new exported API, consider:

- can this stay internal?
- is the naming obvious?
- does this increase mental overhead?
- does this make orchestration harder to understand?

Maestro currently prefers:
- explicit lifecycle APIs
- predictable execution flow
- small surface area

over highly abstracted APIs.

---

# Discussions and Ideas

Discussions and feature ideas are welcome.

Current areas of interest:

- persistence adapters
- async orchestration
- retries and timers
- workflow versioning
- observability
- developer experience improvements

---

# Security

If you discover a security issue, please avoid opening a public issue immediately.

Instead, contact the maintainer privately first.

---

# License

By contributing, you agree that your contributions will be licensed under the MIT License.
