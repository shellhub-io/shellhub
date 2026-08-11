-- Three indexes from 001 that no query in either repo can reach, maintained on every session
-- insert and update for nothing. IF EXISTS because operators may already have dropped them by
-- hand, and an error here fails the boot.
--
-- There is no read-path trade to weigh: these back no filter, no sort and no constraint, so
-- nothing gets slower.

-- The session list accepts exactly three filter fields (device_uid, closed, active) and
-- rejects anything else at the route, and its sort is hardcoded to started_at — username is
-- reachable by neither.
DROP INDEX IF EXISTS sessions_username_idx;

--bun:split

-- Same for type. The type predicates that do exist are all on session_events, a different
-- table, and are already served by its session_id/seat indexes.
DROP INDEX IF EXISTS sessions_type_idx;

--bun:split

-- Redundant rather than merely unused: closed is a near-constant (a session is closed for all
-- but the few minutes it is live), so (closed, started_at) is a strictly fatter duplicate of
-- sessions_started_at_idx and loses to it on cost for every shape that exists — including the
-- one it looks tailor-made for, the recording-conversion worker's
-- "WHERE closed AND recorded AND NOT converted ORDER BY started_at DESC". If that worker ever
-- shows up as a cost, the index it wants is a partial one on the backlog alone, not this.
DROP INDEX IF EXISTS sessions_closed_started_idx;
