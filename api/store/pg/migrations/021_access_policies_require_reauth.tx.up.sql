-- Per-policy re-authentication: access granted by a policy with require_reauth
-- set still triggers a fresh per-session re-authentication, even for an
-- already-enrolled key. reauth_period is the freshness window in seconds (null
-- or 0 means "always"): the re-auth is skipped when the identity re-authed
-- within it.
ALTER TABLE access_policies ADD COLUMN IF NOT EXISTS require_reauth boolean NOT NULL DEFAULT false;

--bun:split

ALTER TABLE access_policies ADD COLUMN IF NOT EXISTS reauth_period bigint;
