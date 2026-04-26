DROP INDEX IF EXISTS idx_services_type;
DROP INDEX IF EXISTS idx_services_owner;

ALTER TABLE services
    DROP COLUMN IF EXISTS environment_ids,
    DROP COLUMN IF EXISTS pipeline_ids,
    DROP COLUMN IF EXISTS tags,
    DROP COLUMN IF EXISTS owner,
    DROP COLUMN IF EXISTS language,
    DROP COLUMN IF EXISTS service_type;
