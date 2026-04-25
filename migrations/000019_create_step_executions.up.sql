CREATE TABLE IF NOT EXISTS step_executions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    pipeline_run_id UUID NOT NULL REFERENCES pipeline_runs(id) ON DELETE CASCADE,
    task_run_name VARCHAR(200) NOT NULL,
    step_name VARCHAR(200) NOT NULL,
    step_index INTEGER NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    image_ref TEXT,
    image_digest TEXT,
    command_args JSONB NOT NULL DEFAULT '{}'::jsonb,
    env_public JSONB NOT NULL DEFAULT '{}'::jsonb,
    secret_refs JSONB NOT NULL DEFAULT '[]'::jsonb,
    params_resolved JSONB NOT NULL DEFAULT '{}'::jsonb,
    workspace_mounts JSONB NOT NULL DEFAULT '[]'::jsonb,
    resources JSONB NOT NULL DEFAULT '{}'::jsonb,
    tekton_results JSONB NOT NULL DEFAULT '[]'::jsonb,
    output_digests JSONB NOT NULL DEFAULT '[]'::jsonb,
    log_ref JSONB NOT NULL DEFAULT '{}'::jsonb,
    trace_id VARCHAR(200),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    duration_ms BIGINT,
    exit_code INTEGER,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_step_executions_run_task_step
    ON step_executions(pipeline_run_id, task_run_name, step_name);
CREATE INDEX IF NOT EXISTS idx_step_executions_pipeline_run_id
    ON step_executions(pipeline_run_id);
CREATE INDEX IF NOT EXISTS idx_step_executions_finished_at
    ON step_executions(finished_at);
CREATE INDEX IF NOT EXISTS idx_step_executions_started_at_null_finished
    ON step_executions(started_at) WHERE finished_at IS NULL;