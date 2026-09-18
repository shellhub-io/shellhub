-- An SSH identity is a credential, and until now the only thing that could hold one was a person.
-- An automation is an API key, so a key must be able to hold one too.
--
-- Two nullable columns rather than an owner_type and an owner_id, because a foreign key needs a
-- column whose every value has one referent table. Keeping real foreign keys is what buys the
-- lifecycle: revoking a key removes the keys it enrolled here, in the database, rather than in
-- application code somebody can forget to call. The CHECK is what stops a row owned by both or
-- by neither.
--
-- The foreign key is composite, matching what 028 did for the member: it pins the identity to the
-- key's own namespace instead of trusting the caller to check.
--
-- The two existing foreign keys on user_id, from 016 and 028, are deliberately left alone. A
-- foreign key with a null column is satisfied without being checked, so a row owned by a key
-- passes both. That reads like an oversight and is not; the migration test inserts such a row.
SET LOCAL lock_timeout = '60s';

--bun:split

ALTER TABLE ssh_identities
    ALTER COLUMN user_id DROP NOT NULL;

--bun:split

ALTER TABLE ssh_identities
    ADD COLUMN api_key_id uuid;

--bun:split

ALTER TABLE ssh_identities
    ADD CONSTRAINT ssh_identities_owner_shape
    CHECK (num_nonnulls(user_id, api_key_id) = 1);

--bun:split

ALTER TABLE ssh_identities
    ADD CONSTRAINT ssh_identities_api_key_fkey
    FOREIGN KEY (api_key_id, namespace_id)
    REFERENCES api_keys (id, namespace_id) ON DELETE CASCADE;
