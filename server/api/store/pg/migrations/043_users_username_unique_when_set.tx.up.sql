ALTER TABLE users DROP CONSTRAINT users_username_key;

--bun:split

CREATE UNIQUE INDEX users_username_key ON users USING btree (username) WHERE username <> '';
