SET LOCAL lock_timeout = '60s';

--bun:split

ALTER TABLE sessions
    DROP CONSTRAINT sessions_api_key_id_fkey;

--bun:split

ALTER TABLE sessions
    DROP COLUMN api_key_id;
