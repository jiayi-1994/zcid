CREATE TABLE IF NOT EXISTS health_signals (
    id VARCHAR(255) PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL REFERENCES projects(id),
    target_type VARCHAR(40) NOT NULL,
    target_id VARCHAR(255) NOT NULL,
    source VARCHAR(80) NOT NULL,
    status VARCHAR(20) NOT NULL,
    severity VARCHAR(20) NOT NULL DEFAULT 'info',
    reason VARCHAR(255) NOT NULL DEFAULT '',
    message TEXT NOT NULL DEFAULT '',
    observed_value JSONB NOT NULL DEFAULT '{}',
    observed_at TIMESTAMP NOT NULL,
    stale_after TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_health_signals_project_target
ON health_signals(project_id, target_type, target_id, observed_at DESC);

CREATE INDEX IF NOT EXISTS idx_health_signals_source
ON health_signals(project_id, source, observed_at DESC);

CREATE INDEX IF NOT EXISTS idx_health_signals_stale_after
ON health_signals(stale_after)
WHERE stale_after IS NOT NULL;
