ALTER TABLE connections ADD COLUMN IF NOT EXISTS device_uid character varying NOT NULL DEFAULT '';
