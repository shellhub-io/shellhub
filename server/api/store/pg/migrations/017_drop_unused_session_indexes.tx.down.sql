-- Restores the index set from 001.
CREATE INDEX IF NOT EXISTS sessions_username_idx ON sessions USING btree (username);

--bun:split

CREATE INDEX IF NOT EXISTS sessions_type_idx ON sessions USING btree (type);

--bun:split

CREATE INDEX IF NOT EXISTS sessions_closed_started_idx ON sessions USING btree (closed, started_at);
