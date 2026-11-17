DELETE FROM session_events
WHERE type = 'env';

--bun:split

UPDATE session_events
SET data = ''
WHERE type NOT IN (
        'pty-output',
        'pty-req',
        'window-change',
        'shell',
        'exec',
        'subsystem',
        'signal',
        'exit-status',
        'exit-signal',
        'direct-tcpip',
        'auth-agent-req@openssh.com'
    )
  AND data IS NOT NULL
  AND data <> '';
