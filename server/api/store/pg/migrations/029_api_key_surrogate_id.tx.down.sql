SET LOCAL lock_timeout = '60s';

--bun:split

ALTER TABLE api_keys
    DROP CONSTRAINT api_keys_id_namespace_id_unique;

--bun:split

ALTER TABLE api_keys
    DROP CONSTRAINT api_keys_id_unique;

--bun:split

ALTER TABLE api_keys
    DROP COLUMN id;
