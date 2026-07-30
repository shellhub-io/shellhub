DROP TRIGGER IF EXISTS users_cleanup_invitations ON users;

--bun:split

DROP FUNCTION IF EXISTS cleanup_user_invitations();

--bun:split

DROP TABLE IF EXISTS membership_invitations;

--bun:split

DROP TABLE IF EXISTS user_invitations;

--bun:split

DROP TYPE IF EXISTS membership_invitation_status;
