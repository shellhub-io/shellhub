-- Discriminate install keys by a `type` column instead of the `system` boolean + magic `name`. The
-- two auto-managed system keys per namespace (legacy, pairing) are told apart by `type`; user keys are
-- `user`. This also adds the pairing key: the source attributed to devices accepted through the
-- tenant-less pairing-code flow, so they no longer attribute to the legacy key.
--
-- Fix-forward only, in two respects:
--   1. Existing pairing devices already stamped with the legacy key are byte-identical to keyless-legacy
--      devices and cannot be re-attributed reliably, so they are left as-is.
--   2. The pairing INSERT below relies on the pre-existing (namespace_id, name) unique index. Unlike
--      'legacy' (reserved since migration 013), 'pairing' was never a reserved name, so a namespace that
--      already holds a user key named 'pairing' keeps it and does NOT get a pairing system key (ON
--      CONFLICT DO NOTHING skips it). Code pairing there still works (acceptPairingDevice accepts the
--      device explicitly, independent of the resolved key); it just falls back to keyless-legacy
--      attribution. New namespaces are unaffected: the key is created at NamespaceCreate, before any
--      user key can claim the name. Accepted as deliberate fix-forward, not worth mutating user data for.

ALTER TABLE install_keys ADD COLUMN type text NOT NULL DEFAULT 'user';

-- Backfill the existing legacy system key (migration 013 created one per namespace with system = true,
-- name = 'legacy'). Only system keys are touched; a user key can't be named 'legacy' (it would have
-- collided with that key on creation).
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
