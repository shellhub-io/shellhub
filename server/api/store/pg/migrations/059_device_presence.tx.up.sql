SET LOCAL lock_timeout = '60s';

--bun:split

LOCK TABLE devices IN ACCESS EXCLUSIVE MODE;

--bun:split

CREATE TABLE device_presence (
    device_id character varying PRIMARY KEY REFERENCES devices (id) ON DELETE CASCADE,
    last_seen timestamp with time zone NOT NULL,
    disconnected_at timestamp with time zone
) WITH (fillfactor = 50);

--bun:split

INSERT INTO device_presence (device_id, last_seen, disconnected_at)
SELECT id, last_seen, disconnected_at FROM devices;

--bun:split

ALTER TABLE devices DROP COLUMN last_seen, DROP COLUMN disconnected_at;
