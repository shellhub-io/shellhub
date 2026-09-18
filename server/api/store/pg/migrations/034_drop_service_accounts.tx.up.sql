-- Service accounts are gone: the automation is the API key, and an SSH identity is a credential
-- it owns. The species of the principal used to be written twice, in users.type and again in
-- memberships.role, and that duplication produced four defects in two days.
--
-- Destructive, knowingly. No stable release ever had service accounts, but v0.27.0-rc.3 through
-- rc.11 did, so a release candidate or a development database holds real rows. Each one loses
-- its membership and its enrolled keys. Re-enrolment is POST
-- /api/namespaces/api-key/:name/ssh-identities, against an API key. The down migration restores
-- the enum value and the column; it cannot resurrect a row.
--
-- Postgres cannot drop a value from an enum, so membership_role is rebuilt without 'service'
-- and every column carrying the type moves across before the old one goes. access_policies
-- joined that set in 028, and missing it would fail the final DROP TYPE on a dependency.
--
-- users.type goes with them. Its only reader was the question this removes, so what is left is
-- a column that can hold one value and a type nobody can explain.
SET LOCAL lock_timeout = '60s';

--bun:split

DELETE FROM access_policies WHERE subject_role = 'service';

--bun:split

DELETE FROM membership_invitations WHERE role = 'service';

--bun:split

DELETE FROM users WHERE type = 'service';

--bun:split

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

ALTER TABLE access_policies
    ALTER COLUMN subject_role TYPE membership_role USING subject_role::text::membership_role;

--bun:split

DROP TYPE membership_role_old;

--bun:split

ALTER TABLE users DROP COLUMN type;

--bun:split

DROP TYPE user_type;
