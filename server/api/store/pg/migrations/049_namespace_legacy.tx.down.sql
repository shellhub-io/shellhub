ALTER TABLE namespaces
    ADD COLUMN ssh_access_mode text NOT NULL DEFAULT 'identity',
    ADD COLUMN ssh_legacy_allowed boolean NOT NULL DEFAULT false;

--bun:split

UPDATE namespaces SET ssh_access_mode = 'legacy', ssh_legacy_allowed = true WHERE legacy;

--bun:split

ALTER TABLE namespaces
    ADD CONSTRAINT namespaces_ssh_access_mode_check
    CHECK (ssh_access_mode IN ('legacy', 'identity')),
    ADD CONSTRAINT namespaces_legacy_requires_permission
    CHECK (ssh_access_mode <> 'legacy' OR ssh_legacy_allowed);

--bun:split

ALTER TABLE namespaces DROP COLUMN legacy;
