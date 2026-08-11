-- RESET drops the per-table override, returning both tables to whatever the server-level
-- autovacuum_*_scale_factor is set to. It does not restore a previous per-table value, because
-- there was none to restore.
ALTER TABLE sessions RESET (
    autovacuum_vacuum_scale_factor,
    autovacuum_analyze_scale_factor
);

--bun:split

ALTER TABLE session_events RESET (
    autovacuum_vacuum_scale_factor,
    autovacuum_analyze_scale_factor
);
