DROP INDEX CONCURRENTLY IF EXISTS devices_namespace_created_at_idx;

--bun:split

CREATE INDEX CONCURRENTLY IF NOT EXISTS devices_namespace_created_at_idx
    ON devices (namespace_id, created_at, id);
