ALTER TABLE namespaces ADD COLUMN require_verified_device_key boolean NOT NULL DEFAULT false;

--bun:split

ALTER TABLE provisioning_key_events
    ADD COLUMN refusal text,
    ADD COLUMN attempts integer NOT NULL DEFAULT 1;

--bun:split

ALTER TABLE systems DROP COLUMN authentication_device_require_verified_key;
