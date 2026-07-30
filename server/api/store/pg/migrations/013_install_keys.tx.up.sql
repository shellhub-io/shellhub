-- Install keys: reusable, revocable, namespace-scoped device-registration credentials. The whole
-- feature lands in one migration (the incremental 013-028 series was never released, so it was folded
-- here). This also drops the namespace-wide device_auto_accept setting from migration 003: a valid
-- install key is now the accept, so the global switch is redundant.
CREATE TABLE install_keys (
    key_digest character(64) NOT NULL,
    namespace_id uuid NOT NULL,
    name character varying NOT NULL,
    reusable boolean DEFAULT false NOT NULL,
    usage_limit integer DEFAULT 0 NOT NULL,
    used_times integer DEFAULT 0 NOT NULL,
    last_used_at timestamp with time zone,
    ephemeral boolean DEFAULT false NOT NULL,
    ephemeral_timeout integer DEFAULT 10 NOT NULL,
    tags text[] DEFAULT '{}' NOT NULL,
    revoked boolean DEFAULT false NOT NULL,
    disabled boolean DEFAULT false NOT NULL,
    system boolean DEFAULT false NOT NULL,
    key_encrypted text,
    key_hint varchar(16),
    mode text DEFAULT 'automatic' NOT NULL,
    webhook_url text,
    webhook_secret text,
    allowed_macs text[] DEFAULT '{}' NOT NULL,
    webhook_timeout integer DEFAULT 0 NOT NULL,
    webhook_callback_ttl integer DEFAULT 0 NOT NULL,
    user_id uuid NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone,
    PRIMARY KEY (key_digest, namespace_id),
    FOREIGN KEY (namespace_id) REFERENCES namespaces(id) ON DELETE CASCADE,
    CONSTRAINT install_keys_mode_check CHECK (mode IN ('automatic', 'manual', 'allowlist', 'webhook'))
);

--bun:split

CREATE UNIQUE INDEX install_keys_namespace_id_name_unique ON install_keys USING btree (namespace_id, name);

--bun:split

-- Enrollment bookkeeping on the device: whether it is ephemeral and its offline timeout (copied from
-- the key at enrollment), the digest of the key it enrolled with, and when its policy was last
-- (re-)evaluated (for reconcile throttling).
ALTER TABLE devices ADD COLUMN IF NOT EXISTS ephemeral boolean DEFAULT false NOT NULL;

--bun:split

ALTER TABLE devices ADD COLUMN IF NOT EXISTS ephemeral_timeout integer DEFAULT 0 NOT NULL;

--bun:split

ALTER TABLE devices ADD COLUMN IF NOT EXISTS install_key_id character(64);

--bun:split

ALTER TABLE devices ADD COLUMN IF NOT EXISTS last_enrollment_attempt_at timestamptz;

--bun:split

-- One legacy (system) key per existing namespace: the source of devices that enroll with only a tenant
-- ID. Manual mode, so those land pending; agents never present it (it is resolved by the system flag).
INSERT INTO install_keys (key_digest, namespace_id, name, reusable, usage_limit, used_times, ephemeral, tags, revoked, system, mode, user_id, created_at, updated_at, expires_at)
SELECT encode(sha256(('system:' || id::text)::bytea), 'hex'), id, 'legacy', true, 0, 0, false, '{}', false, true, 'manual', owner_id, now(), now(), NULL
FROM namespaces
ON CONFLICT DO NOTHING;

--bun:split

-- Attribute every existing device to its namespace's legacy key.
UPDATE devices d
SET install_key_id = sk.key_digest
FROM install_keys sk
WHERE sk.namespace_id = d.namespace_id AND sk.system = true;

--bun:split

-- Now that every device points to a real key, enforce it. install_key_id is nullzero (NULL when a
-- device has no key), and a composite FK skips NULL rows (MATCH SIMPLE), so this never blocks a device.
ALTER TABLE devices ADD CONSTRAINT devices_install_key_fkey
    FOREIGN KEY (install_key_id, namespace_id) REFERENCES install_keys(key_digest, namespace_id) ON DELETE SET NULL (install_key_id);

--bun:split

-- Append-only enrollment history: one immutable row per device that enrolls with a key. Device facts
-- are denormalized so the audit survives a later rename or removal (ephemeral included), and the
-- decision is frozen on the row (decided_status/decided_at) so it outlives the device too. Rows cascade
-- away only with the key or namespace.
CREATE TABLE install_key_events (
    id uuid NOT NULL,
    install_key_id character(64) NOT NULL,
    namespace_id uuid NOT NULL,
    device_uid character(64) NOT NULL,
    hostname character varying NOT NULL,
    mac character varying,
    info_id text,
    info_pretty_name character varying,
    info_version character varying,
    info_arch character varying,
    info_platform character varying,
    public_key text,
    source_ip character varying,
    ephemeral boolean DEFAULT false NOT NULL,
    re_registration boolean DEFAULT false NOT NULL,
    decided_status text,
    decided_at timestamptz,
    created_at timestamp with time zone NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (namespace_id) REFERENCES namespaces(id) ON DELETE CASCADE,
    FOREIGN KEY (install_key_id, namespace_id) REFERENCES install_keys(key_digest, namespace_id) ON DELETE CASCADE,
    CONSTRAINT install_key_events_decided_status_check CHECK (decided_status IN ('accepted', 'rejected'))
);

--bun:split

-- The history list orders a key's events newest-first.
CREATE INDEX install_key_events_key_time_idx ON install_key_events USING btree (namespace_id, install_key_id, created_at DESC);

--bun:split

-- The decision stamp and reconcile look up a device's newest event by device_uid.
CREATE INDEX install_key_events_device_time_idx ON install_key_events USING btree (device_uid, created_at DESC);

--bun:split

-- Single-use ledger for deferred-decision webhook callback tokens (by JWT id): a replayed callback URL
-- finds its jti present and is refused. Pruned by the EnrollmentCallbackCleanup cron past the max TTL.
CREATE TABLE enrollment_callback_redemptions (
    jti text PRIMARY KEY,
    redeemed_at timestamptz NOT NULL
);

--bun:split

CREATE INDEX enrollment_callback_redemptions_redeemed_at_idx ON enrollment_callback_redemptions USING btree (redeemed_at);

--bun:split

-- Install keys replace the namespace-wide auto-accept: a valid key is now the accept.
ALTER TABLE namespaces DROP COLUMN IF EXISTS device_auto_accept;
