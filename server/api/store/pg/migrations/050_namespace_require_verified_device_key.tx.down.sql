ALTER TABLE systems ADD COLUMN authentication_device_require_verified_key boolean NOT NULL DEFAULT false;

--bun:split

ALTER TABLE systems ALTER COLUMN authentication_device_require_verified_key SET DEFAULT true;

--bun:split

ALTER TABLE provisioning_key_events
    DROP COLUMN attempts,
    DROP COLUMN refusal;

--bun:split

ALTER TABLE namespaces DROP COLUMN require_verified_device_key;
