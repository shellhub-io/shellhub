ALTER TABLE provisioning_key_events
    ADD COLUMN replaced_device_uid TEXT,
    ADD COLUMN replaced_device_name TEXT,
    ADD CONSTRAINT provisioning_key_events_refusal_or_replacement CHECK (refusal IS NULL OR replaced_device_uid IS NULL);

ALTER TABLE provisioning_keys
    ADD COLUMN replace_offline_devices BOOLEAN NOT NULL DEFAULT true;
