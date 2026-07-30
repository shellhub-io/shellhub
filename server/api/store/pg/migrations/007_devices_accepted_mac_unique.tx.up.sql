-- Enforce at most one accepted device per (namespace_id, mac). A device's uid is
-- derived from its public key, so reinstalling the agent registers a new row with
-- the same MAC; without this two accepted rows can coexist and a connect resolving
-- by hostname may pick the offline one ("no connection").

-- Step a: keep the most-recently-seen accepted device per (namespace_id, mac) and
-- demote the rest. removed_at is set so the DeviceCleanup cron can later purge them
-- (it filters removed_at < now-30d); the namespace counters are kept in sync.
WITH winners AS (
    SELECT DISTINCT ON (namespace_id, mac) id
    FROM devices
    WHERE status = 'accepted'
    ORDER BY namespace_id, mac, last_seen DESC, id ASC
),
demoted AS (
    UPDATE devices
    SET status = 'removed',
        status_updated_at = now(),
        removed_at = now()
    WHERE status = 'accepted'
      AND id NOT IN (SELECT id FROM winners)
    RETURNING namespace_id
),
counts AS (
    SELECT namespace_id, count(*) AS n
    FROM demoted
    GROUP BY namespace_id
)
UPDATE namespaces ns
SET devices_accepted_count = ns.devices_accepted_count - c.n,
    devices_removed_count  = ns.devices_removed_count  + c.n
FROM counts c
WHERE ns.id = c.namespace_id;

--bun:split

-- Step b: only accepted rows are constrained; pending/rejected/removed are free.
CREATE UNIQUE INDEX devices_accepted_mac_unique
    ON devices USING btree (namespace_id, mac)
    WHERE status = 'accepted';
