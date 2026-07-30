-- The identity SSH access mode: instead of a namespace-wide key ACL plus
-- firewall rules, every SSH login resolves to a ShellHub account and is
-- authorized by Access Policies. A key the namespace already knows connects
-- straight through; an unknown one is held at the gateway until a member
-- approves it in the console.
--
-- Existing namespaces are grandfathered: they stay in the legacy mode and keep
-- the two-way toggle. New namespaces are born in identity and can never go
-- back, which is what the second statement encodes: the column defaults apply
-- to rows created from here on, while the first statement backfills the rows
-- that already exist.
ALTER TABLE namespaces
    ADD COLUMN ssh_access_mode text NOT NULL DEFAULT 'legacy',
    ADD COLUMN ssh_legacy_allowed boolean NOT NULL DEFAULT true;

--bun:split

ALTER TABLE namespaces
    ALTER COLUMN ssh_access_mode SET DEFAULT 'identity',
    ALTER COLUMN ssh_legacy_allowed SET DEFAULT false;

--bun:split

-- Bind a session to the ShellHub user who authorized it via browser approval.
-- NULL for password/public-key logins and web-terminal sessions, which have no
-- approver. ON DELETE SET NULL so removing a user never deletes session history.
ALTER TABLE sessions ADD COLUMN user_id uuid;

--bun:split

ALTER TABLE sessions ADD CONSTRAINT sessions_user_id_fkey
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

--bun:split

-- Access Policies: namespace-scoped, default-deny authorization grants for the
-- identity access mode. A policy grants a subject (user, role, or all members)
-- access to the devices selected by its filter, as the listed logins.
--
-- require_reauth makes access granted by this policy trigger a fresh per-session
-- re-authentication even for an already-enrolled key. reauth_period is the
-- freshness window in seconds (null or 0 means "always"): the re-auth is skipped
-- when the identity re-authed within it.
CREATE TABLE access_policies (
    id uuid NOT NULL,
    namespace_id uuid NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    name character varying NOT NULL,
    subject_type character varying NOT NULL,
    subject_value character varying DEFAULT ''::character varying NOT NULL,
    filter_hostname character varying DEFAULT ''::character varying,
    logins text[] NOT NULL DEFAULT '{}'::text[],
    action character varying NOT NULL DEFAULT 'allow',
    source_ip text[] NOT NULL DEFAULT '{}'::text[],
    require_reauth boolean NOT NULL DEFAULT false,
    reauth_period bigint,
    PRIMARY KEY (id),
    FOREIGN KEY (namespace_id) REFERENCES namespaces(id) ON DELETE CASCADE
);

--bun:split

CREATE INDEX access_policies_namespace_id ON access_policies USING btree (namespace_id);

--bun:split

CREATE TABLE access_policy_tags (
    access_policy_id uuid NOT NULL,
    tag_id uuid NOT NULL,
    created_at timestamp with time zone NOT NULL,
    PRIMARY KEY (access_policy_id, tag_id),
    FOREIGN KEY (access_policy_id) REFERENCES access_policies(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

--bun:split

CREATE INDEX access_policy_tags_access_policy_id ON access_policy_tags USING btree (access_policy_id);

--bun:split

CREATE INDEX access_policy_tags_tag_id ON access_policy_tags USING btree (tag_id);

--bun:split

-- SSH Identities: bindings of an SSH public key to a ShellHub account within a
-- namespace. A connection whose presented key's fingerprint resolves here is
-- recognized as the bound account, with no browser step. Distinct from
-- public_keys (the legacy namespace-scoped key ACL), which is untouched.
CREATE TABLE ssh_identities (
    id uuid NOT NULL,
    namespace_id uuid NOT NULL,
    user_id uuid NOT NULL,
    -- SSH public key fingerprint, "SHA256:…" form (50 chars); varchar to fit it.
    fingerprint character varying NOT NULL,
    data bytea NOT NULL,
    name character varying NOT NULL DEFAULT '',
    -- How this identity came to exist:
    --
    --   'manual'   a public key pasted on the SSH Identities page
    --   'browser'  the web terminal's own key, generated in the browser and held
    --              non-extractably in IndexedDB, so it can never be presented
    --              from anywhere else and dies with the browser's site data
    --   'approval' a key accepted at login, after a native client offered one
    --              the namespace did not know yet
    --
    -- The default is deliberately the weakest claim, which is also what makes
    -- pruning possible later: an unused 'browser' row is almost certainly a key
    -- that no longer exists anywhere, while an unused pasted key still has a
    -- private half sitting on somebody's disk.
    source character varying NOT NULL DEFAULT 'manual'
        CHECK (source IN ('manual', 'browser', 'approval')),
    created_at timestamp with time zone NOT NULL,
    last_used_at timestamp with time zone,
    -- When this identity last completed a re-authentication (Access Policy
    -- require_reauth). Distinct from last_used_at, which is stamped on every
    -- connect; this only moves on a successful re-auth, so it can gate the
    -- reauth_period freshness window.
    last_reauth_at timestamp with time zone,
    -- Key lifecycle. Human identities normally leave these at their defaults
    -- (never-expiring, reusable); any identity may carry a TTL, and a service
    -- account may also make its key single-use. A key is dead once expired or
    -- consumed.
    -- expires_at: null = never expires.
    expires_at timestamp with time zone,
    -- single_use: the key authorizes exactly one established SSH session.
    single_use boolean NOT NULL DEFAULT false,
    -- consumed_at: stamped when a single_use key is burned by its one session.
    consumed_at timestamp with time zone,
    PRIMARY KEY (id),
    -- A fingerprint maps to exactly one identity within a namespace; the same key
    -- may still enroll in other namespaces.
    UNIQUE (namespace_id, fingerprint),
    FOREIGN KEY (namespace_id) REFERENCES namespaces(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

--bun:split

CREATE INDEX ssh_identities_namespace_id ON ssh_identities USING btree (namespace_id);

--bun:split

CREATE INDEX ssh_identities_user_id ON ssh_identities USING btree (user_id);

--bun:split

-- Service accounts: non-human principals that hold an SSH identity for automated
-- systems (CI, backups, config management). A service account is a users row with
-- type='service' plus a namespace membership; it never signs in to the console and
-- is not an API principal. type is the discriminator (not the role) so it survives
-- a future roles->groups migration.
CREATE TYPE user_type AS ENUM (
    'human',
    'service'
);

--bun:split

ALTER TABLE users ADD COLUMN type user_type NOT NULL DEFAULT 'human';

--bun:split

-- Service accounts get a dedicated membership role so human-role policies never match
-- them (roles are matched by exact equality) and a role=service subject targets all of
-- them at once.
ALTER TYPE membership_role ADD VALUE IF NOT EXISTS 'service';

--bun:split

-- SSH login approvals: a decision the SSH gateway parks while it holds a
-- pure-OpenSSH login open, and a member resolves in the console. The short code
-- is printed in the terminal banner and is the row's identity.
--
-- Two kinds, both acting on the identity rather than on the session (the session
-- is only registered once the auth pipeline clears): 'identity' binds the
-- presented key as a new SSH identity, 'reauth' refreshes the re-auth window of
-- one that already exists, because an access policy demanded a fresh one.
--
-- Rows are short-lived. Expiry is expires_at, not a delete: a lookup past it is
-- treated as unknown, and a cron prunes the leftovers.
CREATE TABLE ssh_approvals (
    -- The correlation code from pkg/pairingcode, 8 chars of the Crockford
    -- alphabet. It is the secret: whoever holds it can read and decide the
    -- request, subject to namespace membership.
    code character varying NOT NULL,
    namespace_id uuid NOT NULL,
    -- 'identity' or 'reauth'.
    kind character varying NOT NULL,
    -- The login being decided on, captured at request time so the console can
    -- render it without re-resolving anything.
    session_uid character varying NOT NULL,
    sshid character varying NOT NULL,
    device_uid character varying NOT NULL,
    device_name character varying NOT NULL DEFAULT '',
    username character varying NOT NULL,
    ip_address character varying NOT NULL,
    -- The presented SSH public key: its fingerprint is shown to the user, and
    -- data is what an 'identity' approval persists as the new identity.
    fingerprint character varying NOT NULL,
    data bytea NOT NULL,
    -- The policy's re-auth window in seconds, on a 'reauth' approval, so the
    -- console can say how long confirming lasts. Null means it asks every time.
    reauth_period integer,
    -- 'pending', 'confirmed' or 'rejected'. The transition out of 'pending' is
    -- what makes a decision single-use: it is claimed by affected-row count.
    state character varying NOT NULL DEFAULT 'pending',
    -- Who decided. For an 'identity' approval this is the account the key binds
    -- to, and the gateway adopts it as the session's identity.
    decided_by uuid,
    requested_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL,
    PRIMARY KEY (code),
    FOREIGN KEY (namespace_id) REFERENCES namespaces(id) ON DELETE CASCADE,
    FOREIGN KEY (decided_by) REFERENCES users(id) ON DELETE CASCADE
);

--bun:split

CREATE INDEX ssh_approvals_expires_at ON ssh_approvals USING btree (expires_at);
