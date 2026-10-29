DROP INDEX CONCURRENTLY IF EXISTS session_events_session_type_seat_idx;

--bun:split

CREATE INDEX CONCURRENTLY IF NOT EXISTS session_events_session_type_seat_idx
    ON session_events (session_id, type, seat);
