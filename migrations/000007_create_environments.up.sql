-- Story 3.2: environments table for environment management
CREATE TABLE IF NOT EXISTS environments (
    id VARCHAR(255) PRIMARY KEY,
    project_id VARCHAR(255) NOT NULL REFERENCES projects(id),
    name VARCHAR(100) NOT NULL,
    namespace VARCHAR(100) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS uk_environments_namespace ON environments(namespace) WHERE status != 'deleted';
CREATE UNIQUE INDEX IF NOT EXISTS uk_environments_project_name ON environments(project_id, name) WHERE status != 'deleted';
CREATE INDEX IF NOT EXISTS idx_environments_project_id ON environments(project_id);
