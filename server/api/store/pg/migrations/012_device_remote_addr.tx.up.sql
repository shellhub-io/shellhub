ALTER TABLE devices
    ADD COLUMN IF NOT EXISTS remote_addr character varying(64) NOT NULL DEFAULT '';
