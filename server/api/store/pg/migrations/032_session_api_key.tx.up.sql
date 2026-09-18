-- A session opened by an automation has no account to attribute it to: sessions.user_id has a
-- foreign key to users, and an API key's id there fails the insert, so the session would never
-- start. The key gets a column of its own, the same pair of nullable foreign keys the SSH
-- identity uses.
--
-- No CHECK here, unlike ssh_identities: a session with neither is ordinary. Legacy sessions
-- and device-initiated ones have no principal at all.
--
-- The foreign key is added NOT VALID on purpose. Validating it reads every row of what is
-- already the largest table, and migrations run inline at startup, so that scan would be the
-- boot. The column is null on every existing row, so there is nothing to validate; the
-- constraint is enforced on every insert and update from here on either way.
SET LOCAL lock_timeout = '60s';

--bun:split

ALTER TABLE sessions
    ADD COLUMN api_key_id uuid;

--bun:split

ALTER TABLE sessions
    ADD CONSTRAINT sessions_api_key_id_fkey
    FOREIGN KEY (api_key_id) REFERENCES api_keys (id) ON DELETE SET NULL NOT VALID;
