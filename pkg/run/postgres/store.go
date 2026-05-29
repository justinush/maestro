package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/justinush/maestro/pkg/run"
)

// Store implements [run.Store] with Postgres (state as JSONB).
type Store struct {
	pool *pgxpool.Pool
}

var _ run.Store = (*Store)(nil)

// NewStore returns a Store backed by pool.
//
// The caller owns pool lifecycle. Apply [ApplySchema] before first use.
func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// Create inserts a new run. rec.RunID must be set; stored revision becomes 1.
func (s *Store) Create(ctx context.Context, rec *run.RunRecord) error {
	if rec == nil || rec.RunID == "" {
		return fmt.Errorf("run: invalid record")
	}
	stored, err := cloneRecord(rec)
	if err != nil {
		return err
	}
	stored.Revision = 1

	stateJSON, err := json.Marshal(stored.State)
	if err != nil {
		return fmt.Errorf("postgres: marshal state: %w", err)
	}

	tag, err := s.pool.Exec(ctx, `
		INSERT INTO workflow_runs (run_id, workflow_id, workflow_version, revision, state)
		VALUES ($1, $2, $3, $4, $5::jsonb)
	`, stored.RunID, stored.WorkflowID, stored.WorkflowVersion, stored.Revision, stateJSON)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return run.ErrExists
		}
		return fmt.Errorf("postgres: insert: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("postgres: insert: no rows")
	}
	return nil
}

// Get returns a deep copy of the run or [run.ErrNotFound].
func (s *Store) Get(ctx context.Context, runID string) (*run.RunRecord, error) {
	if runID == "" {
		return nil, run.ErrNotFound
	}
	var (
		rec       run.RunRecord
		stateJSON []byte
	)
	err := s.pool.QueryRow(ctx, `
		SELECT run_id, workflow_id, workflow_version, revision, state
		FROM workflow_runs
		WHERE run_id = $1
	`, runID).Scan(&rec.RunID, &rec.WorkflowID, &rec.WorkflowVersion, &rec.Revision, &stateJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, run.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("postgres: get: %w", err)
	}
	if err := json.Unmarshal(stateJSON, &rec.State); err != nil {
		return nil, fmt.Errorf("postgres: unmarshal state: %w", err)
	}
	return cloneRecord(&rec)
}

// Save updates the run when rec.Revision matches the stored value (then bumps revision).
func (s *Store) Save(ctx context.Context, rec *run.RunRecord) error {
	if rec == nil || rec.RunID == "" {
		return run.ErrNotFound
	}
	stored, err := cloneRecord(rec)
	if err != nil {
		return err
	}

	stateJSON, err := json.Marshal(stored.State)
	if err != nil {
		return fmt.Errorf("postgres: marshal state: %w", err)
	}

	tag, err := s.pool.Exec(ctx, `
		UPDATE workflow_runs
		SET workflow_id = $2,
		    workflow_version = $3,
		    revision = revision + 1,
		    state = $4::jsonb,
		    updated_at = now()
		WHERE run_id = $1 AND revision = $5
	`, stored.RunID, stored.WorkflowID, stored.WorkflowVersion, stateJSON, stored.Revision)
	if err != nil {
		return fmt.Errorf("postgres: save: %w", err)
	}
	if tag.RowsAffected() == 1 {
		return nil
	}

	var exists bool
	err = s.pool.QueryRow(ctx, `SELECT true FROM workflow_runs WHERE run_id = $1`, stored.RunID).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return run.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("postgres: save check: %w", err)
	}
	return run.ErrRevisionConflict
}

func cloneRecord(r *run.RunRecord) (*run.RunRecord, error) {
	if r == nil {
		return nil, nil
	}
	b, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}
	var out run.RunRecord
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return &out, nil
}
