-- Partial on pending: it is the only status the install key list's count reads, and pending is a
-- small, short-lived slice of a namespace's devices. Migration 020 dropped the indexes on
-- last_seen and disconnected_at because a presence heartbeat writes them and an indexed column
-- changing disqualifies HOT; neither is in this index, and status changes only on a decision.
--
-- lock_timeout because migrations run inline at startup before the listener binds, and a lock
-- queued behind a long snapshot would stall the boot. SET LOCAL is enough: bun runs a .tx. file
-- in one transaction.
SET LOCAL lock_timeout = '60s';

--bun:split

CREATE INDEX IF NOT EXISTS devices_pending_by_install_key
    ON devices USING btree (namespace_id, install_key_id)
    WHERE status = 'pending';
