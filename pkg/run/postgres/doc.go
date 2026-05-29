// Package postgres provides a Postgres implementation of [github.com/justinush/maestro/pkg/run.Store].
//
// Workflow run state is stored in the workflow_runs table (JSONB column state).
// Canonical DDL lives in schema.sql ([SchemaDDL] returns the same bytes).
//
// The table and column names are intended to remain stable across v0.x releases for this adapter.
//
// # Usage
//
//	pool, err := pgxpool.New(ctx, databaseURL)
//	if err != nil { ... }
//	store := postgres.NewStore(pool)
//
// [Store] implements [run.Store]: Create, Get, Save with optimistic locking via revision.
// Store is safe for concurrent use when backed by a shared *pgxpool.Pool.
//
// # Schema management
//
// [Store] expects workflow_runs to exist. Two supported approaches:
//
//   - [ApplySchema] - optional helper for examples, tests, and quick local setup (idempotent).
//   - Your migration tool - copy schema.sql or use [SchemaDDL] in goose / golang-migrate / Atlas / etc.
//
// # Scope
//
// This package persists workflow runs only. Application-owned tables (users, applicants, etc.)
// remain the embedder's responsibility and are not coordinated in a single transaction with
// [run.Store] operations in v0.1.
package postgres
