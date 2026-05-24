// Package run provides persistence types for workflow runs.
//
// Store [RunRecord] values with [Store] (for example [MemoryStore] in tests).
// Build records from a live instance with [RecordFromInstance], then restore with
// [github.com/justinush/maestro/pkg/maestro].Runtime.RestoreInstance (preferred) or
// [InstanceFromRecord] for custom [Store] integrations.
//
// [Store.Save] uses optimistic locking via [RunRecord.Revision]; stale writes return [ErrRevisionConflict].
package run
