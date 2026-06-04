# Workflow versioning design (v0.2.x)

How to ship new workflow YAML without breaking runs that are already in flight.

This doc is about **policy** — who picks which version when, and what we intentionally keep out of the engine. Registry mechanics (`pkg/workflow`, `RunRecord`) are in [architecture](../architecture.md#workflow-registry) and [README — multiple workflows](../../README.md#multiple-workflows).

**Status:** design draft. No new Maestro APIs are promised here until we agree on this text and validate it in a real host integration (production-style embed or an internal spike).

---

## Goals

- Make versioning rules explicit before we add more API surface.
- Keep **restore** boring: always the same `workflowId` + `workflowVersion` that were stored when the run started.
- Keep **“what version do new users get?”** in the host — config, flags, a small catalog — not hidden inside the engine.
- Stay embeddable: no workflow database, no publishing service, no “latest” magic on resume.

## Non-goals (for this pass)

- A hosted workflow store or admin UI for publishing YAML.
- Automatic migration of snapshot state from version A to B.
- Forcing in-flight runs onto a newer definition (unless the host builds that on purpose later).
- Semver enforcement or compatibility matrices inside Maestro.

---

## The rules

These are the contracts we want hosts and operators to rely on.

1. **Restore is exact.** `RunRecord` carries `workflowId` and `workflowVersion`. Resume looks up that pair in the registry. Maestro does not substitute “latest” on restore.

2. **New runs use an explicit key.** The host chooses a `workflow.Key{ID, Version}` (directly or via a catalog) and calls `Registry.NewInstance`. The engine does not guess.

3. **No silent upgrades.** If a host deploys `kyc.sg.main@1.1.0`, existing runs on `1.0.0` keep running and restoring as `1.0.0` until they finish or the host handles them explicitly.

4. **Published versions are immutable.** Same `(id, version)` must not change meaning in place. Registering the same key twice fails (`ErrDuplicateKey`). In production, ship a **new** version string when the graph changes; don’t edit the artifact for an already-released version.

5. **“Active” / “default” version is host policy.** Which version new SG MAIN starts use comes from the host routing table, feature flag, or `VersionCatalog` — not from `RestoreInstance`.

---

## Concepts (three layers)

There are three separate layers. Mixing them up is where versioning bugs come from.

| Layer | What it is | Who owns it |
|-------|------------|-------------|
| **Registry** | Loaded definitions: `kyc.sg.main@1.0.0`, `kyc.sg.main@1.1.0`, … | Maestro `workflow.Registry` at process startup |
| **Catalog** | Policy: “for new SG MAIN runs, use version **1.1.0**” | Host app (config, code, remote config) |
| **Run record** | Fact: “this run **is** `kyc.sg.main` **1.0.0**” | Host store (e.g. Postgres), set at create time |

```text
Registry (all versions still needed to run or restore)
  kyc.sg.main@1.0.0
  kyc.sg.main@1.1.0

Catalog (host — only for *starting* runs)
  kyc.sg.main → active version 1.1.0

RunRecord (immutable for the life of the run)
  workflowId: kyc.sg.main
  workflowVersion: 1.0.0
```

The registry answers: *“Is this definition loaded?”*  
The catalog answers: *“Which version should **new** runs use?”*  
The run record answers: *“Which version is **this** run?”*

Maestro today implements the registry + run record wiring. The catalog is **host** code.

---

## Lifecycle (typical rollout)

1. **Publish** — Add `kyc.sg.main` version `1.1.0` as new YAML (new file or new version field). Register both `1.0.0` and `1.1.0` in the same process (e.g. `workflow.LoadDir`).

2. **Activate for new runs** — Point the catalog (or route map) at `1.1.0`. New `POST /kyc/start` (or equivalent) paths resolve to `workflow.Key{ID: "kyc.sg.main", Version: "1.1.0"}`.

3. **Drain old runs** — In-flight `1.0.0` runs keep restoring `1.0.0`. No engine step changes their stored version.

4. **Retire a version from deploy** — Once nothing in the DB still needs `1.0.0`, stop loading it. Restores for old records will get `ErrNotFound` until that version is loaded again (or those runs are archived). That’s operational discipline, not a special Maestro feature.

**Rollback** is the same pattern in reverse: set the catalog back to `1.0.0` for **new** starts. Runs already on `1.1.0` stay on `1.1.0` unless the host builds a migration path.

---

## How this maps to `pkg/workflow`

Already shipped (v0.2 registry work):

- `workflow.Key` = YAML `id` + `version`, same fields as `RunRecord`.
- `Registry.Register` / `LoadDir` — multiple keys, including multiple versions per `id`.
- `Registry.NewInstance(key, …)` — host passes the key for new runs.
- `Registry.RestoreInstance(rec, …)` — lookup `(rec.WorkflowID, rec.WorkflowVersion)`; mismatch guard via `ErrDefinitionMismatch`.

What we are **not** adding in the first versioning slice:

- `ResolveLatest(id)`
- `RegisterDefault(id, version)`
- Updating `RunRecord.WorkflowVersion` inside Maestro

Optional later helpers (only if a real host repeats the same boilerplate):

- `Registry.ListVersions(id []string)` — discover what’s loaded
- A tiny **example** `VersionCatalog` type in `examples/` or this doc, not in core

---

## Host pattern: catalog + registry

A common host pattern folds policy into one map — entity + flow → full `workflow.Key` with a pinned version:

```go
{"SG", "MAIN"}: {ID: "kyc.sg.main", Version: "1.0.0"},
```

That’s already a catalog; it just hard-codes the version next to the route.

When shipping `1.1.0`, we split policy from routing in code (one map or two steps):

```go
// Policy: which version is "active" for new runs on this workflow id
func activeVersion(workflowID string) string {
    switch workflowID {
    case "kyc.sg.main":
        return "1.1.0" // was 1.0.0 before rollout
    default:
        return "1.0.0"
    }
}

// Route still picks workflow id; catalog picks version for NEW runs only
key := workflow.Key{ID: "kyc.sg.main", Version: activeVersion("kyc.sg.main")}
reg.NewInstance(key, opts)
```

Restore never calls `activeVersion`:

```go
reg.RestoreInstance(rec, opts) // uses rec.WorkflowID + rec.WorkflowVersion only
```

Business routing (`entity` + `flow` → workflow **id**) stays separate from version policy (**id** → active **version** for starts).

---

## KYC-style scenarios

Sanity check for Binance-scale KYC hosts. Expected behavior under this design:

| Scenario | Host action | Effect on runs |
|----------|-------------|----------------------|
| **New country** | New workflow **id** (e.g. `kyc.br.main@1.0.0`) + new route row | New market, new id — not a version bump on `kyc.sg.main` |
| **Add POA step to SG main** | New **version** `1.1.0`, same id; catalog sends new starts to `1.1.0` | In-flight stays on `1.0.0` until done |
| **Rollback bad 1.1.0** | Catalog back to `1.0.0` for new starts; keep `1.1.0` in registry while drains finish | New users get `1.0.0`; existing `1.1.0` runs unchanged |
| **Old user resumes** | Nothing special — `Get(runID)` → `RestoreInstance(rec)` | Always `1.0.0` if that’s what’s on the record |
| **Periodic refresh** | Different workflow **id** (`kyc.sg.refresh`) | Version policy on `kyc.sg.main` doesn’t apply; routing picks another id |

If a scenario needs “move everyone on 1.0.0 to 1.1.0 mid-run,” that’s a **host-defined migration** (new run, manual step, or a future tool). Out of scope for v0.2.x.

---

## Snapshot compatibility (be honest)

Pinning `workflowVersion` on restore only guarantees the **same graph** as at start. It does **not** guarantee that a snapshot taken under `1.0.0` is safe to resume under `1.1.0` after steps, variables, or guards change.

Rule of thumb for hosts:

- **Patch-level** changes (copy, timeouts) — often fine when the graph structure is unchanged.
- **Graph changes** (new step, renamed step id, different variables) — treat as a new version; don’t expect Maestro to merge old state into a new definition.

Cross-version resume is a future “migration” topic, not something we pretend works today.

---

## Version strings

Hosts can use whatever string fits their release process. Semver (`1.1.0`) is a good human convention; Maestro treats `version` as an opaque label.

- Bump **version** when the workflow **definition** that new runs should use changes.
- Don’t confuse with YAML `schemaVersion` (Maestro’s document shape) — that’s about the file format, not which business flow a user is on.

---

## Operations checklist

- Load every `(id, version)` pair that still has runs in the DB (or accept restore failures).
- Never rewrite in-place YAML for a version already shipped.
- Deploy catalog changes (active version) independently of draining old runs.
- Monitor `ErrNotFound` on restore — usually “forgot to load that version” or “retired too early.”

---

## MVP deliverables (after this doc is agreed)

Smallest useful slice — no workflow DB:

| Deliverable | Notes |
|-------------|--------|
| This design doc | Public in `docs/design/` |
| Short pointer in `docs/architecture.md` / roadmap | Link here |
| `examples/` snippet | Optional second YAML `1.1.0` + catalog comment or tiny `VersionCatalog` |
| Maestro API changes | **Probably none** for first merge; revisit `ListVersions` only if real host code needs it |

Explicitly **not** in MVP: publishing API, Postgres workflow table, engine “latest” resolver.

---

## Open questions

Things to settle in review, not blockers for writing this doc:

1. **Deprecation metadata** — Do we want optional YAML fields (`deprecated: true`) for lint/docs only, or keep it entirely in host config?
2. **When to remove old versions from `LoadDir`** — Time-based drain vs count of open runs — host ops playbook.
3. **Feature-flagged catalog** — Same binary loads all versions; flag picks active — enough for KYC studio?
4. **Public example** — Show `1.0.0` + `1.1.0` side by side in `examples/` without forcing every embedder to adopt two files on day one.

---

## Related docs

- [Architecture — workflow registry](../architecture.md#workflow-registry)
- [README — multiple workflows](../../README.md#multiple-workflows)
- [Roadmap](../roadmap.md) — versioning listed under “Next”

---

## Summary

- **Registry** holds all loaded `(id, version)` pairs.
- **Catalog** (host) picks the active version for **new** runs only.
- **RunRecord** stores the exact version forever; restore never auto-upgrades.
- Rollouts and rollbacks are catalog + ops, not engine magic.
- v0.2.x ships clarity (docs + examples) before any workflow publishing store.
