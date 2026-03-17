CREATE TABLE IF NOT EXISTS pipeline_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    pipeline_id UUID NOT NULL REFERENCES pipelines(id),
    project_id VARCHAR(255) NOT NULL REFERENCES projects(id),
    run_number INT NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    trigger_type VARCHAR(20) NOT NULL DEFAULT 'manual',
    triggered_by VARCHAR(255) REFERENCES users(id),
    git_branch VARCHAR(200),
    git_commit VARCHAR(40),
    git_author VARCHAR(200),
    git_message TEXT,
    config_snapshot JSONB NOT NULL DEFAULT '{}',
    params JSONB DEFAULT '{}',
    tekton_name VARCHAR(200),
    namespace VARCHAR(100),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    error_message TEXT,
    artifacts JSONB DEFAULT '[]',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uk_pipeline_runs_number ON pipeline_runs(pipeline_id, run_number);
CREATE INDEX IF NOT EXISTS idx_pipeline_runs_pipeline ON pipeline_runs(pipeline_id);
CREATE INDEX IF NOT EXISTS idx_pipeline_runs_project ON pipeline_runs(project_id);
CREATE INDEX IF NOT EXISTS idx_pipeline_runs_status ON pipeline_runs(status);
