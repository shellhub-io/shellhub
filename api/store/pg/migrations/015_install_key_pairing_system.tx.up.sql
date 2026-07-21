-- Discriminate install keys by a `type` column instead of the `system` boolean + magic `name`. The
-- two auto-managed system keys per namespace (legacy, pairing) are told apart by `type`; user keys are
-- `user`. This also adds the pairing key: the source attributed to devices accepted through the
-- tenant-less pairing-code flow, so they no longer attribute to the legacy key.
--
-- Fix-forward only: existing pairing devices already stamped with the legacy key are byte-identical to
-- keyless-legacy devices and cannot be re-attributed reliably, so they are left as-is.

ALTER TABLE install_keys ADD COLUMN type text NOT NULL DEFAULT 'user';

-- Backfill the existing legacy system key (migration 013 created one per namespace with system = true,
-- name = 'legacy'). Only system keys are touched; a user key can't be named 'legacy' — it would have
-- collided with that key on creation.
UPDATE install_keys SET type = 'legacy' WHERE system = true AND name = 'legacy';

-- One pairing (system) key per existing namespace. Automatic mode (acceptance is the code itself);
-- agents never present it (resolved by type). Digest is deterministic per namespace.
INSERT INTO install_keys (key_digest, namespace_id, name, reusable, usage_limit, used_times, ephemeral, tags, revoked, system, type, mode, user_id, created_at, updated_at, expires_at)
SELECT encode(sha256(('system:pairing:' || id::text)::bytea), 'hex'), id, 'pairing', true, 0, 0, false, '{}', false, true, 'pairing', 'automatic', owner_id, now(), now(), NULL
FROM namespaces
ON CONFLICT DO NOTHING;

ALTER TABLE install_keys DROP COLUMN system;

-- Guarantee exactly one legacy and one pairing key per namespace; user keys are unconstrained here.
CREATE UNIQUE INDEX install_keys_namespace_type_unique ON install_keys (namespace_id, type) WHERE type <> 'user';
