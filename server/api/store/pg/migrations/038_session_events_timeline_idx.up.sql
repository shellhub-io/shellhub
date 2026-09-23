DROP INDEX CONCURRENTLY IF EXISTS session_events_timeline_idx;

--bun:split

CREATE INDEX CONCURRENTLY IF NOT EXISTS session_events_timeline_idx
    ON session_events (session_id, created_at, id) WHERE type <> 'pty-output';
