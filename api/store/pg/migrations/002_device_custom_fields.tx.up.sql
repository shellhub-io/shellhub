ALTER TABLE devices ADD COLUMN IF NOT EXISTS custom_fields jsonb NOT NULL DEFAULT '{}';
