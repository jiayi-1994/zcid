-- Story 2.3: Casbin policy persistence table
CREATE TABLE IF NOT EXISTS public.casbin_rule (
    id BIGSERIAL PRIMARY KEY,
    ptype VARCHAR(16) NOT NULL,
    v0 VARCHAR(255),
    v1 VARCHAR(255),
    v2 VARCHAR(255),
    v3 VARCHAR(255),
    v4 VARCHAR(255),
    v5 VARCHAR(255)
);

CREATE INDEX IF NOT EXISTS idx_casbin_rule_ptype ON public.casbin_rule (ptype);
CREATE INDEX IF NOT EXISTS idx_casbin_rule_v0 ON public.casbin_rule (v0);
CREATE INDEX IF NOT EXISTS idx_casbin_rule_v1 ON public.casbin_rule (v1);
