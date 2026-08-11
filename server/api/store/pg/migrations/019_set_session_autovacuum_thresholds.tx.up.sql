-- Autovacuum's default scale factors are proportions of the table, so the bigger a table
-- gets the longer it waits: at the default vacuum factor of 0.2, a 6.7M-row session_events
-- does not get vacuumed until ~1.3M rows are dead, and at the default analyze factor of 0.1
-- its statistics are only refreshed after ~670k changes. Measured on a production instance,
-- that left session_events 45 days stale on a table the planner has to estimate against, and
-- sessions carrying 24k dead tuples ten days after its last vacuum.
--
-- The proportion is the wrong shape for these two tables specifically: they are the largest
-- in the database and the fastest growing, which is exactly the combination the default
-- punishes. Lowering the factors converts "a fifth of the table" into a bound that stays
-- workable as the table grows — ~335k dead rows before a vacuum and ~134k changes before an
-- analyze at today's size, instead of 1.3M and 670k.
--
-- These are storage parameters, not a rewrite: the ALTER takes a brief SHARE UPDATE EXCLUSIVE
-- lock, touches only the catalog and cannot block reads or writes. It also does not vacuum or
-- analyze anything by itself — it only changes when the daemon next decides to, which is why
-- the ANALYZE below is needed to clear the backlog the old factor already allowed.
--
-- Retention (the cron this migration ships alongside) reduces how much dead tuple churn these
-- tables see at all, but does not replace this: a prune is itself a large source of dead
-- tuples, so the thresholds matter more once it is running, not less.
ALTER TABLE sessions SET (
    autovacuum_vacuum_scale_factor = 0.05,
    autovacuum_analyze_scale_factor = 0.02
);

--bun:split

ALTER TABLE session_events SET (
    autovacuum_vacuum_scale_factor = 0.05,
    autovacuum_analyze_scale_factor = 0.02
);

--bun:split

-- Statistics on the instance this was measured on were 45 days stale, and the lowered factor
-- above does nothing about a backlog that already exists — it only brings the *next* analyze
-- forward. The planner estimates against these tables on every session list, so leaving them
-- stale is the part that actually produces bad plans.
--
-- Unlike a VACUUM FULL this is safe to run at boot: ANALYZE reads a bounded random
-- sample (default_statistics_target * 300, so ~30k rows) rather than the whole table, takes
-- only SHARE UPDATE EXCLUSIVE, and unlike VACUUM it is allowed inside a transaction block.
ANALYZE sessions;

--bun:split

ANALYZE session_events;
