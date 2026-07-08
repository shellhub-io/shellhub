-- Bind a session to the ShellHub user who authorized it via browser approval.
-- NULL for password/public-key logins and web-terminal sessions, which have no
-- approver. ON DELETE SET NULL so removing a user never deletes session history.
ALTER TABLE sessions ADD COLUMN IF NOT EXISTS user_id uuid;
ALTER TABLE sessions ADD CONSTRAINT sessions_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;
