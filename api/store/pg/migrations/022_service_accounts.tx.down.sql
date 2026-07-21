ALTER TABLE users DROP COLUMN IF EXISTS type;

--bun:split

DROP TYPE IF EXISTS user_type;

--bun:split

-- Postgres cannot drop a single enum value, so rebuild membership_role without
-- 'service'. The column cast fails on purpose if any membership still uses it,
-- which would mean down-migrating with live service accounts.
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

DROP TYPE membership_role_old;
