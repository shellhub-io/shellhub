SET LOCAL lock_timeout = '60s';

--bun:split

LOCK TABLE devices IN ACCESS EXCLUSIVE MODE;

--bun:split

ALTER TABLE devices
    ADD COLUMN last_seen timestamp with time zone NOT NULL DEFAULT 'epoch',
    ADD COLUMN disconnected_at timestamp with time zone;

--bun:split

UPDATE devices AS device
SET last_seen = presence.last_seen, disconnected_at = presence.disconnected_at
FROM device_presence AS presence
WHERE presence.device_id = device.id;

--bun:split

ALTER TABLE devices ALTER COLUMN last_seen DROP DEFAULT;

--bun:split

DROP TABLE device_presence;
