# Custom action types (host embedding)

How host applications register app-owned `onEnter` / `onExit` action runners and reference those types in workflow YAML.

Maestro ships built-in action types `stub` and `http`. Hosts often need more — outbound vendor session create, internal service calls, domain-specific side effects. Runtime dispatch already works via `engine.Registry.Register` / `MustRegister`; validation must opt in so YAML cannot reference arbitrary type strings without the host's knowledge.

This doc is the follow-up to [async callbacks bridge validation](async-callbacks.md#bridge-validation-spike-outcome): the spike showed create steps were awkward when validation only allowed `stub` and `http` in YAML.

**Status:** `validate.Options.AllowedActionTypes` ships in v0.2.x. No new execution primitive — only schema/validate ergonomics.

---

## Goals

- Let hosts express app-owned side effects in workflow YAML (`type: vendor-create-session`).
- Keep built-in `stub` and `http` strictly validated in Go.
- Fail fast at workflow load when YAML references types the host has not allowed.
- Match runtime behavior: unregistered types still fail at execution with `engine.ErrUnknownActionType`.

## Non-goals (for this pass)

- Per-type JSON Schema for custom `params` in core (runners validate at execution time).
- Engine-side registry of allowed types (host passes an explicit allowlist).
- Async action runners that pause mid-step.
- Replacing the [async callbacks bridge](async-callbacks.md) — custom actions make the **create** step honest in YAML; resume still uses `SubmitInput` on a blocking input step.

---

## Two registries (do not confuse them)

| Registry | Package | Maps |
|----------|---------|------|
| **Workflow registry** | `pkg/workflow` | `workflow id` + `version` → validated definition |
| **Action registry** | `pkg/engine` | `action type` string → `ActionRunner` |

`workflow.LoadDir` validates YAML. `engine.Registry` executes `onEnter` / `onExit` at runtime. Both must agree on custom type names.

---

## Recipe

### 1. Implement a runner

Implement `engine.ActionRunner` for your type. The runner receives step context, action id, and `params`; it may mutate variables or call external systems.

### 2. Register at startup

```go
actionReg := engine.NewRegistry()
actionReg.MustRegister("stub", engine.NewStubRunner())
actionReg.MustRegister("http", engine.NewHTTPRunner(client))
actionReg.MustRegister("vendor-create-session", vendorCreateRunner)
```

### 3. Allow the type at load / validate time

```go
reg, err := workflow.LoadDir("workflows", validate.Options{
    AllowedActionTypes: []string{
        "vendor-create-session",
    },
})
```

The same allowlist applies to `maestro.LoadWithValidate`, `validate.WorkflowFile`, and `validate.WorkflowDefinition`.

### 4. Reference in YAML

```yaml
- id: create-verification
  kind: action
  onEnter:
    - type: vendor-create-session
      id: create-vendor
      params:
        provider: sumsub
```

### 5. Execute with the same action registry

Pass `ActionRegistry: actionReg` on `maestro.InstanceOptions` for both `NewInstance` and `RestoreInstance`.

```go
in, err := reg.NewInstance(key, maestro.InstanceOptions{
    RunID:          runID,
    ActionRegistry: actionReg,
})
```

---

## Validation rules

| Rule | Behavior |
|------|----------|
| Built-in types | `stub`, `http` always valid; param shape validated strictly in Go |
| Custom types | Must appear in `validate.Options.AllowedActionTypes` |
| Type name shape | kebab-case (`vendor-create-session`) |
| Built-in in allowlist | Error — do not list `stub` or `http` in `AllowedActionTypes` |
| Duplicates in allowlist | Error |
| Unregistered at runtime | `engine.ErrUnknownActionType` when the step runs |
| Custom `params` | Accepted as JSON objects; Maestro does not validate custom param shapes in v1 |

At validate time, Maestro extends the embedded v0.1 JSON Schema `action.type` enum with your allowed types. Built-in enum entries are unchanged.

---

## CLI

```bash
maestro validate -f workflows/kyc.sg.vendor.yaml \
  --allowed-action-type vendor-create-session
```

Repeat `--allowed-action-type` for multiple types. Use the same allowlist in CI as in `workflow.LoadDir` at app startup.

---

## Relationship to async callbacks

The [async callbacks bridge](async-callbacks.md) models vendor waits as:

1. **Action step** — outbound create (idempotent).
2. **Resume step** — `kind: human`, blocks until webhook → `SubmitInput`.

Before custom action validation, hosts often used `type: stub` in YAML and ran create logic outside the graph after `RunUntilBlocked`. With `AllowedActionTypes`, the create step can use `type: vendor-create-session` so the graph matches reality.

Resume semantics are unchanged: webhook → host `externalRef` lookup → `SubmitInput`. See [Callback identity](async-callbacks.md#callback-identity) for the host mapping table contract.

---

## Checklist for host integrators

- [ ] Runner registered on `engine.Registry` before any run.
- [ ] Same type strings in `AllowedActionTypes` and `MustRegister`.
- [ ] CI `maestro validate` uses the same allowlist as production `LoadDir`.
- [ ] Custom runner validates its own `params` and returns clear errors.
- [ ] Idempotent create uses keys such as `runId` + `stepId` (see async callbacks doc).

---

## Related docs

- [Async callbacks design](async-callbacks.md) — bridge pattern and `externalRef` mapping
- [Architecture — pause / resume](../architecture.md#pause-resume-model)
- [README — multiple workflows](../../README.md#multiple-workflows)

---

## Summary

- **Runtime:** `engine.Registry` already dispatches custom action types.
- **Validation:** opt in with `validate.Options.AllowedActionTypes`.
- **Built-ins:** `stub` and `http` unchanged; custom types get structural checks only.
- **Next host step:** replace stub + out-of-band create workarounds with honest YAML action types.
