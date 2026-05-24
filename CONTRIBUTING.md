# Contributing

Thanks for your interest in contributing to Maestro.

Maestro is still in early `v0.x` development. The current focus is keeping the runtime simple, understandable, and easy to embed into backend applications.

Before opening a large feature PR, please open an issue or discussion first.

---

# Development Setup

Requirements:

- Go 1.26+
- Git
- Make (recommended)

Clone the repository:

```bash
git clone https://github.com/justinush/maestro.git
cd maestro
```

Install dev tools (first time):

```bash
make install-tools
```

Run the full local gate (fmt-check, lint, tests, smoke, CLI build):

```bash
make ci
```

GitHub Actions runs golangci-lint in a separate job; the workflow then runs:

```bash
make ci-action
```

That is `verify`, `fmt-check`, `vet`, `test`, `smoke`, and `build` — no `make lint` in that job. The same workflow job then runs **`make test-race`**; if race fails, that job fails.

Maintainers: before a version tag on `main`, run `make release-check` (`verify`, `fmt-check`, `vet`, `test`, `smoke`, `doc-check`). The GitHub release workflow on version tags also runs golangci-lint and builds the binary.

Individual targets:

```bash
make test          # unit tests (-count=1, 10m timeout)
make test-race     # unit tests with -race
make smoke         # CLI validate/simulate scenarios
make doc-check     # pkg examples + go doc sanity for public APIs
make lint          # golangci-lint
make fmt           # gofumpt format
make fmt-check     # fail if gofumpt would change files
make lint-fix      # golangci-lint --fix
make build         # build dist/maestro
```

Quick test only:

```bash
go test ./...
```

Formatting:

- Run **`make fmt`** before a PR (gofumpt, matches extra-rules in `.golangci.yml`).
- If `make lint` reports import order issues, run **`make lint-fix`** or fix imports manually (`gci` in `.golangci.yml`).

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

- run **`make ci`** (or at minimum `make check` and `make smoke`)
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

When changing exported APIs in `pkg/`, update godoc comments and run `make doc-check`. Add or adjust `Example*` tests in `pkg/*_test.go` when the canonical usage path changes.

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
