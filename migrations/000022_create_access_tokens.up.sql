CREATE TABLE IF NOT EXISTS access_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    token_type VARCHAR(20) NOT NULL,
    name VARCHAR(120) NOT NULL,
    token_prefix VARCHAR(24) NOT NULL,
    token_hash VARCHAR(64) UNIQUE NOT NULL,
    scopes TEXT NOT NULL,
    user_id VARCHAR(255) REFERENCES users(id),
    project_id UUID REFERENCES projects(id),
    created_by VARCHAR(255) REFERENCES users(id),
    expires_at TIMESTAMPTZ NOT NULL,
    last_used_at TIMESTAMPTZ,
    revoked_at TIMESTAMPTZ,
    revoked_by VARCHAR(255) REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_access_tokens_type CHECK (token_type IN ('personal', 'project')),
    CONSTRAINT chk_access_tokens_owner CHECK (
        (token_type = 'personal' AND user_id IS NOT NULL AND project_id IS NULL) OR
        (token_type = 'project' AND project_id IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS idx_access_tokens_hash ON access_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_access_tokens_user ON access_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_access_tokens_project ON access_tokens(project_id);
CREATE INDEX IF NOT EXISTS idx_access_tokens_expires ON access_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_access_tokens_revoked ON access_tokens(revoked_at);
