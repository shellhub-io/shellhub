ALTER TABLE namespaces DROP COLUMN IF EXISTS ssh_legacy_allowed;

--bun:split

ALTER TABLE namespaces ALTER COLUMN ssh_access_mode SET DEFAULT 'legacy';
