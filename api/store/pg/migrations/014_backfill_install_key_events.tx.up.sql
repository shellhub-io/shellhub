-- Migration 013 attributed every existing device to its namespace's legacy key (devices.install_key_id)
-- but created no registration-history row for them. The console routes device acceptance through each
-- key's registration activity and the devices list shows only accepted devices, so a device that was
-- pending at upgrade time would be unreachable: absent from the accepted-only list and from every key's
-- activity, with no way to accept it. One synthetic event per device that has a key but no event closes
-- that gap. The decision is frozen for devices already accepted/rejected so the history reads correctly;
-- a pending device gets an open event (decided_status NULL), which is what surfaces the accept control.
INSERT INTO install_key_events (
    id, install_key_id, namespace_id, device_uid, hostname, mac, public_key, source_ip,
    ephemeral, re_registration, decided_status, decided_at, created_at
)
SELECT
    gen_random_uuid(),
    d.install_key_id,
    d.namespace_id,
    d.id,
    d.name,
    d.mac,
    d.public_key,
    d.remote_addr,
    d.ephemeral,
    false,
    CASE WHEN d.status IN ('accepted', 'rejected') THEN d.status::text END,
    CASE WHEN d.status IN ('accepted', 'rejected') THEN d.status_updated_at END,
    d.created_at
FROM devices d
WHERE d.install_key_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM install_key_events e WHERE e.device_uid = d.id);
