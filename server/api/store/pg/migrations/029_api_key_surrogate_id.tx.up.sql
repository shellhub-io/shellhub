-- An API key is about to be something other rows point at: an SSH identity it owns, and an
-- access policy that names it. Neither can point at what the table has today. The primary key is
-- (key_digest, namespace_id), and a foreign key to key_digest would spread credential-derived
-- material through joins, indexes and logs; the name is unique per namespace but UpdateAPIKey
-- changes it, and a rename must not silently move what the key owns.
--
-- So the table gains an id that means nothing and never changes. The default is volatile, which
-- rewrites the table and is what gives each existing row its own value; api_keys is small enough
-- that this is cheaper than a backfill in two steps.
--
-- Both uniques are load-bearing. (id) is what a plain foreign key needs. (id, namespace_id) is
-- what a composite one needs, and a composite one is how a referencing row is pinned to the key's
-- own namespace by the database instead of by the application remembering to check.
SET LOCAL lock_timeout = '60s';

--bun:split

ALTER TABLE api_keys
    ADD COLUMN id uuid NOT NULL DEFAULT gen_random_uuid();

--bun:split

ALTER TABLE api_keys
    ADD CONSTRAINT api_keys_id_unique UNIQUE (id);

--bun:split

ALTER TABLE api_keys
    ADD CONSTRAINT api_keys_id_namespace_id_unique UNIQUE (id, namespace_id);
