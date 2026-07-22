-- Identity-first namespaces: new namespaces are born in the identity SSH access
-- mode and can never switch to legacy. ssh_legacy_allowed marks the namespaces
-- that predate this change (grandfathered): they keep the two-way toggle.
ALTER TABLE namespaces ADD COLUMN IF NOT EXISTS ssh_legacy_allowed boolean NOT NULL DEFAULT false;

--bun:split

-- Every existing namespace is grandfathered.
UPDATE namespaces SET ssh_legacy_allowed = true;

--bun:split

-- Defensive only: the Go store writes ssh_access_mode explicitly, so this default
-- covers raw SQL inserts that omit the column.
ALTER TABLE namespaces ALTER COLUMN ssh_access_mode SET DEFAULT 'identity';
