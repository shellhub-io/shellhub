CREATE TYPE session_type AS ENUM (
    'shell',
    'exec',
    'scp',
    'sftp',
    'subsystem',
    'term',
    'web',
    'heredoc',
    'unknown',
    'none'
);

--bun:split

ALTER TABLE sessions ADD COLUMN type session_type,
                     ADD COLUMN closed boolean NOT NULL DEFAULT false;
