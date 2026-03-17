CREATE TABLE IF NOT EXISTS pipelines (
    id                 UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id         VARCHAR(255) NOT NULL REFERENCES projects(id),
    name               VARCHAR(200) NOT NULL,
    description        TEXT,
    status             VARCHAR(20) NOT NULL DEFAULT 'draft',
    config             JSONB NOT NULL DEFAULT '{}',
    trigger_type       VARCHAR(20) NOT NULL DEFAULT 'manual',
    concurrency_policy VARCHAR(20) NOT NULL DEFAULT 'queue',
    created_by         VARCHAR(255) NOT NULL REFERENCES users(id),
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_pipelines_project_name ON pipelines(project_id, name) WHERE status != 'deleted';
CREATE INDEX IF NOT EXISTS idx_pipelines_project_id ON pipelines(project_id);
CREATE INDEX IF NOT EXISTS idx_pipelines_status ON pipelines(status);
CREATE INDEX IF NOT EXISTS idx_pipelines_trigger_type ON pipelines(trigger_type);
