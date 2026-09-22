-- 016 made legacy access a grandfathered privilege: it backfilled the existing rows with
-- ('legacy', true) and then set the defaults to ('identity', false) for everything created after.
-- The pair was never meant to hold legacy without the permission, but nothing enforced that where
-- rows are written, only in the service that edits the mode. The admin CLI takes the mode from a
-- flag and leaves the permission at Go's zero value, so it produced exactly that pair.
--
-- The UPDATE is not housekeeping. It is what makes the constraint appliable: dev and
-- release-candidate databases already hold rows that would fail it. 016 shipped in v0.27.0-rc.*
-- only, so no stable release has these columns and there is nothing else to repair.
UPDATE namespaces SET ssh_legacy_allowed = true WHERE ssh_access_mode = 'legacy';

--bun:split

-- The mode column has been plain text since 016, so the first constraint closes a typo landing as
-- a silently unrecognised mode. The second gives the invariant an owner that no future writer can
-- forget, which is the whole point: a rule enforced only by the code paths that remember it is a
-- habit, not an invariant.
ALTER TABLE namespaces
    ADD CONSTRAINT namespaces_ssh_access_mode_check
    CHECK (ssh_access_mode IN ('legacy', 'identity')),
    ADD CONSTRAINT namespaces_legacy_requires_permission
    CHECK (ssh_access_mode <> 'legacy' OR ssh_legacy_allowed);
