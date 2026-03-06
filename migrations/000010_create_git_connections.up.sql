CREATE TABLE git_connections (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          VARCHAR(100) NOT NULL,
    provider_type VARCHAR(20)  NOT NULL,
    server_url    VARCHAR(500) NOT NULL,
    access_token  TEXT         NOT NULL,
    refresh_token TEXT,
    token_type    VARCHAR(20)  NOT NULL DEFAULT 'pat',
    status        VARCHAR(20)  NOT NULL DEFAULT 'connected',
    description   TEXT         NOT NULL DEFAULT '',
    created_by    VARCHAR(255) NOT NULL REFERENCES users(id),
    created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX uk_git_connections_name ON git_connections(name) WHERE status != 'deleted';
CREATE INDEX idx_git_connections_provider_type ON git_connections(provider_type);
CREATE INDEX idx_git_connections_status ON git_connections(status);
