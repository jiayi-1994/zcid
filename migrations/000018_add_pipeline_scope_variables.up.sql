-- Add pipeline_id for pipeline-scope variables (FR11)
ALTER TABLE variables ADD COLUMN IF NOT EXISTS pipeline_id VARCHAR(255);
CREATE UNIQUE INDEX IF NOT EXISTS uk_variables_pipeline_key ON variables(project_id, pipeline_id, key) WHERE scope = 'pipeline' AND status != 'deleted';
CREATE INDEX IF NOT EXISTS idx_variables_pipeline ON variables(pipeline_id) WHERE scope = 'pipeline' AND status != 'deleted';
