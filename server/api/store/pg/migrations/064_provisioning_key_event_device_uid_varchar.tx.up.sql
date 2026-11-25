SET LOCAL lock_timeout = '60s';

--bun:split

ALTER TABLE provisioning_key_events ALTER COLUMN device_uid TYPE character varying;
