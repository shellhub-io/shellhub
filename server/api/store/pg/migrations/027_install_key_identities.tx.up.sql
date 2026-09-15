-- Neither column ever held MAC addresses in the field: an agent started with
-- SHELLHUB_PREFERRED_IDENTITY reports whatever the operator chose, commonly a serial number, and
-- that is what the allowlist is compared against.
--
-- A new migration rather than an edit of 013, which created both columns: 013 shipped in
-- v0.27.0-rc.2 and RC images are published and upgraded between, while bun records only a
-- migration's number and no checksum. Editing it in place is a silent no-op on any database that
-- already applied it, leaving the old column name behind until the first query fails.
SET LOCAL lock_timeout = '60s';

--bun:split

ALTER TABLE install_keys RENAME COLUMN allowed_macs TO allowed_identities;

--bun:split

ALTER TABLE install_key_events RENAME COLUMN mac TO identity;
