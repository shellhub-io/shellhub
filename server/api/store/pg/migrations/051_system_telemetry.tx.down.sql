DROP INDEX systems_singleton;

ALTER TABLE systems
    DROP COLUMN telemetry_enabled,
    DROP COLUMN instance_id;
