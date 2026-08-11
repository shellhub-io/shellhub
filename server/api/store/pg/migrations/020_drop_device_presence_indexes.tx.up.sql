-- PostgreSQL disqualifies HOT whenever an indexed column changes, and every device presence
-- heartbeat writes last_seen — so this index is what forced each beat to rewrite the heap
-- tuple and touch every index on the table. The device list's ORDER BY last_seen DESC sorts
-- over a sequential scan instead.
--
-- lock_timeout because migrations run inline in startup before the listener binds: an
-- ACCESS EXCLUSIVE request queued behind a long snapshot would otherwise stall the boot
-- indefinitely. SET LOCAL is enough here because bun runs a .tx. file in one transaction.
SET LOCAL lock_timeout = '60s';

--bun:split

-- IF EXISTS: operators may already have dropped this by hand, and an error fails the boot.
DROP INDEX IF EXISTS devices_last_seen;

--bun:split

-- Unused: nothing filters or orders on disconnected_at alone, only as half of the online
-- predicate, which is too unselective to be worth an index scan.
DROP INDEX IF EXISTS devices_disconnected_at;
