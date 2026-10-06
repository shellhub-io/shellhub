DROP INDEX IF EXISTS users_username_key;

--bun:split

ALTER TABLE users ADD CONSTRAINT users_username_key UNIQUE (username);
