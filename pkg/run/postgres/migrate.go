package postgres

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schemaSQL string

// SchemaDDL returns the canonical workflow_runs DDL (same as schema.sql).
//
// Use this when applying the schema with your own migration tool (goose, golang-migrate, Atlas, etc.).
// Table and column names are intended to remain stable across v0.x releases.
func SchemaDDL() string {
	return schemaSQL
}

// ApplySchema creates the workflow_runs table and indexes (idempotent).
//
// Optional convenience for examples, local development, and integration tests.
// Production applications may prefer [SchemaDDL] with their own migration pipeline instead.
//
// [Store] requires the schema to exist before use; how you apply it is up to the embedder.
func ApplySchema(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return fmt.Errorf("postgres: nil pool")
	}
	if _, err := pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("postgres: apply schema: %w", err)
	}
	return nil
}
