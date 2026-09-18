-- Both principal columns on sessions carry ON DELETE SET NULL, and Postgres does not index a
-- referencing column for you. Without an index, deleting a principal is a sequential scan of
-- sessions, which is the largest table here and the one already under scrutiny for exactly this
-- class of query.
--
-- user_id has been in that state since 016; this fixes it rather than adding a second one
-- beside it. Both indexes are partial because almost every row has neither: a partial index is
-- a fraction of the size and is all the SET NULL needs.
--
-- CREATE INDEX CONCURRENTLY cannot run inside a transaction, which is why this file has no
-- .tx. suffix. 022 is the precedent.
--
-- Each create is preceded by a drop because a concurrent build that is interrupted leaves an
-- index that exists and is invalid. IF NOT EXISTS matches on the name alone, so it would adopt
-- that corpse: maintained on every write, never used by the planner, and reported as applied.
DROP INDEX CONCURRENTLY IF EXISTS sessions_user_id_idx;

--bun:split

CREATE INDEX CONCURRENTLY IF NOT EXISTS sessions_user_id_idx
    ON sessions (user_id) WHERE user_id IS NOT NULL;

--bun:split

DROP INDEX CONCURRENTLY IF EXISTS sessions_api_key_id_idx;

--bun:split

CREATE INDEX CONCURRENTLY IF NOT EXISTS sessions_api_key_id_idx
    ON sessions (api_key_id) WHERE api_key_id IS NOT NULL;
