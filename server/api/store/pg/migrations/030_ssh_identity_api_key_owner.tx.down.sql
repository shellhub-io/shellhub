-- Reverting drops every identity an API key owns: user_id is about to be NOT NULL again and
-- those rows have none. Nothing can recover them; they are re-enrolled, not restored.
SET LOCAL lock_timeout = '60s';

--bun:split

DELETE FROM ssh_identities WHERE api_key_id IS NOT NULL;

--bun:split

ALTER TABLE ssh_identities
    DROP CONSTRAINT ssh_identities_api_key_fkey;

--bun:split

ALTER TABLE ssh_identities
    DROP CONSTRAINT ssh_identities_owner_shape;

--bun:split

ALTER TABLE ssh_identities
    DROP COLUMN api_key_id;

--bun:split

ALTER TABLE ssh_identities
    ALTER COLUMN user_id SET NOT NULL;
