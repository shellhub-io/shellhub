-- Restores the index set from 001, which reinstates the HOT blocker on the heartbeat and so
-- also restores its write amplification.
CREATE INDEX IF NOT EXISTS devices_last_seen ON devices USING btree (last_seen);

--bun:split

CREATE INDEX IF NOT EXISTS devices_disconnected_at ON devices USING btree (disconnected_at);
