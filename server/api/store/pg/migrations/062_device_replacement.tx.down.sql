ALTER TABLE provisioning_keys
    DROP COLUMN replace_offline_devices;

DELETE FROM provisioning_key_events WHERE replaced_device_uid IS NOT NULL;

ALTER TABLE provisioning_key_events
    DROP CONSTRAINT provisioning_key_events_refusal_or_replacement,
    DROP COLUMN replaced_device_uid,
    DROP COLUMN replaced_device_name;
