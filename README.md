<div align="center">

# Maestro—embedded workflow runtime

Human-in-the-loop workflows for KYC and onboarding.

[![Go Reference](https://pkg.go.dev/badge/github.com/justinush/maestro.svg)](https://pkg.go.dev/github.com/justinush/maestro)
[![CI](https://github.com/justinush/maestro/actions/workflows/ci.yml/badge.svg)](https://github.com/justinush/maestro/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/justinush/maestro)](https://goreportcard.com/report/github.com/justinush/maestro)
[![License](https://img.shields.io/github/license/justinush/maestro)](./LICENSE)
[![Release](https://img.shields.io/github/v/release/justinush/maestro)](https://github.com/justinush/maestro/releases)

**[Introduction](#introduction) &nbsp;&nbsp;&bull;&nbsp;&nbsp;**
**[Quick Start](#quick-start) &nbsp;&nbsp;&bull;&nbsp;&nbsp;**
**[Examples](#examples) &nbsp;&nbsp;&bull;&nbsp;&nbsp;**
**[Architecture](#architecture) &nbsp;&nbsp;&bull;&nbsp;&nbsp;**
**[Roadmap](#roadmap) &nbsp;&nbsp;&bull;&nbsp;&nbsp;**
**[Contributing](#contributing)**

</div>

---

## Introduction

Maestro is an embedded workflow orchestration runtime originally designed for KYC and onboarding systems.

It helps backend applications execute long-running flows involving:

- human approval steps
- external vendor actions
- pause/resume execution
- persistence + restore
- workflow-driven orchestration

Unlike orchestration platforms that require separate infrastructure, Maestro runs directly inside your Go application as a library.

Typical use cases:

- KYC onboarding
- manual review systems
- fintech operations
- multi-step onboarding
- backoffice workflows
- approval flows

---

## Quick Start

Install:

```bash
go get github.com/justinush/maestro
```

Minimal example:

```go
package main

import (
	"fmt"

	"github.com/justinush/maestro/pkg/engine"
	"github.com/justinush/maestro/pkg/maestro"
)

func main() {
	rt, err := maestro.Load("workflow.yaml")
	if err != nil {
		panic(err)
	}

	in, err := rt.NewInstance(maestro.InstanceOptions{})
	if err != nil {
		panic(err)
	}

	res := in.RunUntilBlocked()

	switch res.Status {
	case engine.RunBlocked:
		fmt.Println("workflow paused for human input")

	case engine.RunCompleted:
		fmt.Println("workflow completed")

	case engine.RunFailed:
		panic(res.Err)
	}
}
```

Example workflow:

```yaml
schemaVersion: "0.1"

id: onboarding
version: "1.0.0"

initialStepId: collect-profile
terminalStepIds:
  - approved

steps:
  - id: collect-profile
    kind: human

  - id: approved
    kind: end

transitions:
  - from: collect-profile
    to: approved
```

---

## Architecture

Maestro follows an embedded orchestration model.

Your application owns:

- HTTP APIs
- authentication
- business data
- UI
- database models

Maestro owns:

- workflow graph execution
- orchestration lifecycle
- transitions
- human steps
- execution trace

![Architecture](./docs/assets/architecture.png)

---

## Examples

The repository includes progressively more advanced examples.

| Example | Purpose |
|---|---|
| [`library-basic`](./examples/demos/library-basic) | smallest embedding example |
| [`embed-kyc-service`](./examples/demos/embed-kyc-service) | pause/resume + persistence |
| [`http-runner`](./examples/demos/http-runner) | external HTTP actions |
| [`kyc-backend`](./examples/apps/kyc-backend) | realistic backend integration |

---

## Core Concepts

### Runtime

A loaded and validated workflow definition.

```go
rt, err := maestro.Load("workflow.yaml")
```

### Instance

A running workflow execution.

```go
in, err := rt.NewInstance(maestro.InstanceOptions{})
```

### Human Steps

Execution pauses until the application submits input.

```go
sub := in.SubmitInput(...)
```

### Persistence

Workflow runs can be stored and restored across requests.

```go
rec := run.RecordFromInstance(...)
in, err := rt.RestoreInstance(...)
```

### Actions

Workflow steps can execute external logic through registered runners.

```go
reg := engine.RegistryWithHTTP(client)
```

---

## Repository Structure

| Path | Purpose |
|---|---|
| `pkg/definition` | workflow schema + decoding |
| `pkg/validate` | workflow validation |
| `pkg/engine` | workflow runtime engine |
| `pkg/run` | persistence abstractions |
| `pkg/maestro` | high-level embedding APIs |
| `examples/demos` | focused learning examples |
| `examples/apps` | near real-world applications |

---

## Project Status

Maestro is currently in early `v0.x` development.

The core orchestration APIs are usable today, but some APIs may evolve before `v1` stability.

Current focus areas:

- embedding experience
- workflow lifecycle clarity
- persistence model
- orchestration ergonomics
- real-world backend integration

---

## Roadmap

Planned areas:

- Postgres store adapter
- async callback/webhook flows
- retry policies
- workflow versioning semantics
- improved typed variable access
- richer execution observability

Not planned yet:

- hosted SaaS
- distributed runtime cluster
- visual workflow editor

---

## Why Maestro?

Maestro focuses on:

- embedded-first architecture
- readable workflows
- human + system orchestration
- backend-oriented integration
- application-owned control
- simple runtime embedding

The goal is to make workflow orchestration feel natural inside normal backend services.

---

## Contributing

Contributions, feedback, and discussions are welcome.

Helpful links:

- [examples](./examples)
- [architecture notes](./docs/architecture.md)
- [contributing guide](./CONTRIBUTING.md)

---

## License

[MIT](./LICENSE)
