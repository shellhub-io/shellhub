ALTER TABLE namespaces ADD COLUMN legacy boolean NOT NULL DEFAULT false;

--bun:split

UPDATE namespaces SET legacy = (ssh_access_mode = 'legacy');

--bun:split

ALTER TABLE namespaces
    DROP CONSTRAINT namespaces_legacy_requires_permission,
    DROP CONSTRAINT namespaces_ssh_access_mode_check;

--bun:split

ALTER TABLE namespaces
    DROP COLUMN ssh_access_mode,
    DROP COLUMN ssh_legacy_allowed;
