-- Reclaims the bloat 018 stops accumulating: while HOT was impossible every heartbeat left a
-- dead tuple behind, and the device list's default sort now scans that heap sequentially.
-- Running after 018 also avoids rebuilding the two indexes that just went away, and running
-- after 019 means the rewrite lays the pages out at that fillfactor.
--
-- VACUUM cannot run inside a transaction, which is why this file omits the .tx. suffix and
-- keeps every statement in its own --bun:split chunk (see TestNonTransactionalMigrations).
--
-- The timeouts bound a boot, because migrations run inline before the listener binds. bun marks
-- a migration applied before running it, so if one fires the boot fails but the restart skips
-- this migration and comes up with the table merely still bloated — re-run the VACUUM by hand
-- to finish the job. Peer replicas that boot while this holds the migration lock fail outright
-- rather than wait: migrator.Lock inserts a row, it does not block.
SET lock_timeout = '60s';

--bun:split

SET statement_timeout = '10min';

--bun:split

-- ANALYZE because VACUUM FULL does not refresh planner statistics, and 018 just changed which
-- plans are available for this table.
VACUUM (FULL, ANALYZE) devices;

--bun:split

-- SET rather than SET LOCAL above, since there is no transaction to scope it to. The connection
-- returns to the pool unreset, so the timeouts leak into application queries unless undone.
RESET lock_timeout;

--bun:split

RESET statement_timeout;
