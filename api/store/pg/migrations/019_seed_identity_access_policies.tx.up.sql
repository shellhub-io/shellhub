-- Seed a permissive starter Access Policy for every namespace that is already in
-- the identity SSH access mode but has no policies. Without this, a namespace
-- flipped to identity by migration 010 (converted from the old ssh_require_approval
-- boolean) would default-deny every SSH login and lock everyone out. The starter
-- policy grants all members access to all devices as any login; admins tighten it.
-- Idempotent: only namespaces with zero policies are seeded.
INSERT INTO access_policies (id, namespace_id, created_at, updated_at, name, subject_type, subject_value, filter_hostname, logins)
SELECT gen_random_uuid(), n.id, now(), now(), 'Default access', 'all-members', '', '', ARRAY['*']
FROM namespaces n
WHERE n.ssh_access_mode = 'identity'
  AND NOT EXISTS (SELECT 1 FROM access_policies ap WHERE ap.namespace_id = n.id);
