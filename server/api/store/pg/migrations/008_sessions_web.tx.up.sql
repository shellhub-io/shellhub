-- Mark sessions that originated from the web terminal. Origin (web console vs a
-- native SSH client) is not represented by any existing column: the scalar `type`
-- is unused and the channel (shell/exec/scp) lives in session_events.
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS web boolean NOT NULL DEFAULT false;
