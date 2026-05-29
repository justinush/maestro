// Package postgres provides a Postgres implementation of [github.com/justinush/maestro/pkg/run.Store].
//
// Workflow run state is stored in the workflow_runs table (JSONB column state).
// Table DDL is defined in schema.sql and applied with [ApplySchema].
//
// The table and column names are part of the v0.x persistence contract for this adapter.
//
// # Usage
//
//	pool, err := pgxpool.New(ctx, databaseURL)
//	if err != nil { ... }
//	if err := postgres.ApplySchema(ctx, pool); err != nil { ... }
//	store := postgres.NewStore(pool)
//
// store implements [run.Store]: Create, Get, Save with optimistic locking via revision.
// Implementations are safe for concurrent use when backed by a shared *pgxpool.Pool.
//
// # Scope
//
// This package persists workflow runs only. Application-owned tables (users, applicants, etc.)
// remain the embedder's responsibility and are not coordinated in a single transaction with
// [run.Store] operations in v0.1.
package postgres
