CREATE TABLE variables (
    id VARCHAR(255) PRIMARY KEY,
    scope VARCHAR(20) NOT NULL DEFAULT 'project',
    project_id VARCHAR(255) REFERENCES projects(id),
    key VARCHAR(200) NOT NULL,
    value TEXT NOT NULL DEFAULT '',
    var_type VARCHAR(20) NOT NULL DEFAULT 'plain',
    description TEXT NOT NULL DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX uk_variables_global_key ON variables(key) WHERE scope = 'global' AND status != 'deleted';
CREATE UNIQUE INDEX uk_variables_project_key ON variables(project_id, key) WHERE scope = 'project' AND status != 'deleted';
CREATE INDEX idx_variables_project ON variables(project_id) WHERE status != 'deleted';
CREATE INDEX idx_variables_scope ON variables(scope) WHERE status != 'deleted';
