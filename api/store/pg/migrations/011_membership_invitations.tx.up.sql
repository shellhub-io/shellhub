-- The membership_invitations / user_invitations schema moved into core from the cloud schema.
-- On enterprise databases these objects already exist (the pre-v0.27 cloud 001 migration created
-- them), so every statement here is idempotent: it creates the objects on fresh Community installs
-- and reconciles the pre-existing ones on enterprise upgrades. The only addition over the old cloud
-- schema is the `sig` column (+ its partial unique index) and the user-deletion cleanup trigger.

DO $$ BEGIN
    CREATE TYPE membership_invitation_status AS ENUM (
        'pending',
        'accepted',
        'rejected',
        'cancelled'
    );
EXCEPTION WHEN duplicate_object THEN null;
END $$;

--bun:split

CREATE TABLE IF NOT EXISTS user_invitations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    email character varying(254) NOT NULL,
    status character varying(32) DEFAULT 'pending'::character varying NOT NULL,
    invitations bigint DEFAULT 1 NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (email)
);

--bun:split

CREATE TABLE IF NOT EXISTS membership_invitations (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    tenant_id uuid NOT NULL,
    user_id uuid NOT NULL,
    invited_by uuid NOT NULL,
    role membership_role NOT NULL,
    status membership_invitation_status DEFAULT 'pending'::membership_invitation_status NOT NULL,
    status_updated_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone,
    invitations bigint DEFAULT 1 NOT NULL,
    sig varchar(64),
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (tenant_id) REFERENCES namespaces(id) ON DELETE CASCADE,
    FOREIGN KEY (invited_by) REFERENCES users(id) ON DELETE CASCADE
);

--bun:split

-- sig is new in v0.27; add it to pre-existing enterprise tables (a no-op after a fresh create above).
ALTER TABLE membership_invitations ADD COLUMN IF NOT EXISTS sig varchar(64);

--bun:split

CREATE INDEX IF NOT EXISTS membership_invitations_status_idx ON membership_invitations USING btree (status);

--bun:split

CREATE INDEX IF NOT EXISTS membership_invitations_tenant_user_idx ON membership_invitations USING btree (tenant_id, user_id);

--bun:split

CREATE UNIQUE INDEX IF NOT EXISTS membership_invitations_sig_unique ON membership_invitations (sig) WHERE sig IS NOT NULL;

--bun:split

-- When a user is deleted, their invitation rows must go with them. membership_invitations.user_id
-- has no usable FK (it points to a user OR a user_invitation depending on the case), and
-- user_invitations is global per email, so neither cascades on user deletion. Without this a
-- deleted-then-reinvited email stays wedged: the leftover accepted user_invitation blocks
-- re-completion. A trigger covers every deletion path (admin, orphan cleanup, raw store delete).
CREATE OR REPLACE FUNCTION cleanup_user_invitations() RETURNS trigger AS $$
BEGIN
    DELETE FROM membership_invitations WHERE user_id = OLD.id;
    DELETE FROM user_invitations WHERE id = OLD.id;

    RETURN OLD;
END;
$$ LANGUAGE plpgsql;

--bun:split

DROP TRIGGER IF EXISTS users_cleanup_invitations ON users;

--bun:split

CREATE TRIGGER users_cleanup_invitations
    AFTER DELETE ON users
    FOR EACH ROW EXECUTE FUNCTION cleanup_user_invitations();
