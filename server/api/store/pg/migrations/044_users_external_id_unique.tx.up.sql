UPDATE users SET external_id = NULL WHERE external_id IS NOT NULL;

--bun:split

CREATE UNIQUE INDEX users_external_id_key ON users USING btree (external_id) WHERE external_id IS NOT NULL;
