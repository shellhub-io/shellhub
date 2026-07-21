-- Service accounts: non-human principals that hold an SSH identity for automated
-- systems (CI, backups, config management). A service account is a users row with
-- type='service' plus a namespace membership; it never signs in to the console and
-- is not an API principal. type is the discriminator (not the role) so it survives
-- a future roles->groups migration.
CREATE TYPE user_type AS ENUM (
    'human',
    'service'
);

--bun:split

ALTER TABLE users ADD COLUMN IF NOT EXISTS type user_type NOT NULL DEFAULT 'human';

--bun:split

-- Service accounts get a dedicated membership role so human-role policies never match
-- them (roles are matched by exact equality) and a role=service subject targets all of
-- them at once.
ALTER TYPE membership_role ADD VALUE IF NOT EXISTS 'service';
