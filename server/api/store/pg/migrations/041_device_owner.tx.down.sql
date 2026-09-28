DROP INDEX IF EXISTS devices_namespace_owner_idx;

--bun:split

ALTER TABLE devices DROP CONSTRAINT IF EXISTS devices_owner_member_fkey;

--bun:split

ALTER TABLE devices DROP COLUMN IF EXISTS owner_id;
