CREATE TABLE IF NOT EXISTS workflow_runs (
    run_id           TEXT PRIMARY KEY,
    workflow_id      TEXT NOT NULL,
    workflow_version TEXT NOT NULL,
    revision         BIGINT NOT NULL,
    state            JSONB NOT NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_workflow_runs_revision ON workflow_runs (run_id, revision);
