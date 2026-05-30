// Package run provides persistence types for workflow runs.
//
// Use [Store] to persist [RunRecord] values. [MemoryStore] is intended for tests,
// demos, and local examples (it is not durable).
//
// For durable backends, use [github.com/justinush/maestro/pkg/run/postgres.Store]
// (Postgres + JSONB). Ensure workflow_runs exists before use: optional [postgres.ApplySchema],
// or your own migrations ([postgres.SchemaDDL] / schema.sql).
//
// Build records from a live instance with [RecordFromInstance], then restore with
// [github.com/justinush/maestro/pkg/maestro].Runtime.RestoreInstance (preferred) or
// [InstanceFromRecord] for custom [Store] integrations.
//
// [Store.Save] uses optimistic locking via [RunRecord.Revision]; stale writes return [ErrRevisionConflict].
package run
