SET LOCAL lock_timeout = '60s';

--bun:split

ALTER TABLE sessions
    DROP COLUMN term;
