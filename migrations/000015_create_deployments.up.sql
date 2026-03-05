CREATE TABLE deployments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id VARCHAR(255) NOT NULL REFERENCES projects(id),
    environment_id VARCHAR(255) NOT NULL REFERENCES environments(id),
    pipeline_run_id UUID REFERENCES pipeline_runs(id),
    image VARCHAR(500) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','syncing','healthy','degraded','failed','rolled_back')),
    argo_app_name VARCHAR(200),
    sync_status VARCHAR(20),
    health_status VARCHAR(20),
    error_message TEXT,
    deployed_by VARCHAR(255) NOT NULL REFERENCES users(id),
    started_at TIMESTAMPTZ,
    finished_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_deployments_project ON deployments(project_id);
CREATE INDEX idx_deployments_env ON deployments(environment_id);
CREATE INDEX idx_deployments_status ON deployments(status);
