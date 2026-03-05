DROP INDEX IF EXISTS idx_variables_pipeline;
DROP INDEX IF EXISTS uk_variables_pipeline_key;
ALTER TABLE variables DROP COLUMN IF EXISTS pipeline_id;
