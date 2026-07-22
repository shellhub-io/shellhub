ALTER TABLE access_policies DROP COLUMN IF EXISTS reauth_period;

--bun:split

ALTER TABLE access_policies DROP COLUMN IF EXISTS require_reauth;
