SET LOCAL lock_timeout = '60s';

--bun:split

ALTER TABLE sessions DROP COLUMN type, DROP COLUMN closed;

--bun:split

DROP TYPE session_type;
