-- Instance API keys: credentials that authenticate as an instance administrator rather than as a
-- member of a namespace. They live in their own table because api_keys.namespace_id is part of that
-- table's primary key and so cannot be null, and because keeping the two apart is what makes a
-- namespace key structurally incapable of resolving to an administrator.
--
-- expires_at is NOT NULL: unlike a namespace key, an instance key cannot be created without an
-- expiration date.
CREATE TABLE instance_api_keys (
    key_digest character(64) NOT NULL,
    name character varying NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    PRIMARY KEY (key_digest),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

--bun:split

CREATE UNIQUE INDEX instance_api_keys_name_unique ON instance_api_keys USING btree (name);
