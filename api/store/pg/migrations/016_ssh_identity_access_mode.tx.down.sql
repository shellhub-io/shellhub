DROP TABLE IF EXISTS ssh_approvals;

--bun:split

DROP TABLE IF EXISTS ssh_identities;

--bun:split

DROP TABLE IF EXISTS access_policy_tags;

--bun:split

DROP TABLE IF EXISTS access_policies;

--bun:split

ALTER TABLE sessions DROP CONSTRAINT IF EXISTS sessions_user_id_fkey;

--bun:split

ALTER TABLE sessions DROP COLUMN IF EXISTS user_id;

--bun:split

ALTER TABLE namespaces
    DROP COLUMN IF EXISTS ssh_legacy_allowed,
    DROP COLUMN IF EXISTS ssh_access_mode;

--bun:split

ALTER TABLE users DROP COLUMN IF EXISTS type;

--bun:split

DROP TYPE IF EXISTS user_type;

--bun:split

-- Postgres cannot drop a single enum value, so rebuild membership_role without
-- 'service'. Every column carrying the enum has to move to the rebuilt type
-- before the old one can be dropped, or the drop fails on the ones left behind
-- and takes the whole transactional rollback with it. The casts fail on purpose
-- if any row still holds 'service', which would mean down-migrating with live
-- service accounts.
ALTER TYPE membership_role RENAME TO membership_role_old;

--bun:split

CREATE TYPE membership_role AS ENUM (
    'owner',
    'administrator',
    'operator',
    'observer'
);

--bun:split

ALTER TABLE memberships
    ALTER COLUMN role TYPE membership_role USING role::text::membership_role;

--bun:split

ALTER TABLE api_keys
    ALTER COLUMN role TYPE membership_role USING role::text::membership_role;

--bun:split

ALTER TABLE membership_invitations
    ALTER COLUMN role TYPE membership_role USING role::text::membership_role;

--bun:split

DROP TYPE membership_role_old;
