-- Drop the unused "modelist" key from recorded pty-req events.
UPDATE session_events
SET data = (data::jsonb - 'modelist')::text
WHERE type = 'pty-req';
