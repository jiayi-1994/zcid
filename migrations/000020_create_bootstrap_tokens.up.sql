CREATE TABLE IF NOT EXISTS bootstrap_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    token_hash VARCHAR(64) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_bootstrap_tokens_expires ON bootstrap_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_bootstrap_tokens_unused ON bootstrap_tokens(used_at) WHERE used_at IS NULL;
