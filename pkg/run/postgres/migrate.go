package postgres

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed schema.sql
var schemaSQL string

// ApplySchema creates the workflow_runs table and indexes (idempotent).
//
// Call once per database environment before using [Store].
func ApplySchema(ctx context.Context, pool *pgxpool.Pool) error {
	if pool == nil {
		return fmt.Errorf("postgres: nil pool")
	}
	if _, err := pool.Exec(ctx, schemaSQL); err != nil {
		return fmt.Errorf("postgres: apply schema: %w", err)
	}
	return nil
}
