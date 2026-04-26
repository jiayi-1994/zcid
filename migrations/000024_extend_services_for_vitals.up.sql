ALTER TABLE services
    ADD COLUMN IF NOT EXISTS service_type VARCHAR(40) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS language VARCHAR(80) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS owner VARCHAR(120) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS tags JSONB NOT NULL DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS pipeline_ids JSONB NOT NULL DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS environment_ids JSONB NOT NULL DEFAULT '[]';

CREATE INDEX IF NOT EXISTS idx_services_owner
ON services(project_id, owner)
WHERE status != 'deleted' AND owner != '';

CREATE INDEX IF NOT EXISTS idx_services_type
ON services(project_id, service_type)
WHERE status != 'deleted' AND service_type != '';
