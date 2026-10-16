WITH input AS (
    SELECT :'tenant'::uuid AS tenant, :'name'::varchar AS name, :'owner'::varchar AS owner, :'max_devices'::integer AS max_devices, :'announcement'::text AS announcement
), owner AS (
    SELECT id FROM users WHERE username = (SELECT owner FROM input)
), namespace AS (
    INSERT INTO namespaces (id, created_at, updated_at, scope, name, owner_id, max_devices, record_sessions, connection_announcement, legacy)
    SELECT input.tenant, now(), now(), 'team', input.name, owner.id, input.max_devices, true, input.announcement, true
    FROM input, owner
    RETURNING id
), membership AS (
    INSERT INTO memberships (user_id, namespace_id, created_at, updated_at, role)
    SELECT owner.id, namespace.id, now(), now(), 'owner'
    FROM owner, namespace
), legacy_key AS (
    INSERT INTO provisioning_keys (key_digest, namespace_id, name, mode, reusable, type, user_id, created_at, updated_at)
    SELECT encode(sha256(('system:' || namespace.id)::bytea), 'hex'), namespace.id, 'legacy', 'manual', true, 'legacy', owner.id, now(), now()
    FROM owner, namespace
), pairing_key AS (
    INSERT INTO provisioning_keys (key_digest, namespace_id, name, mode, reusable, type, user_id, created_at, updated_at)
    SELECT encode(sha256(('system:pairing:' || namespace.id)::bytea), 'hex'), namespace.id, 'pairing', 'automatic', true, 'pairing', owner.id, now(), now()
    FROM owner, namespace
)
SELECT id FROM namespace;
