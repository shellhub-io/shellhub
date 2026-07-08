-- Per-namespace SSH authorization mode. "legacy" keeps the key/firewall model;
-- "identity" gates every SSH login on an out-of-band browser approval (no device
-- credential required) and governs access through Access Policies.
ALTER TABLE namespaces ADD COLUMN IF NOT EXISTS ssh_access_mode TEXT NOT NULL DEFAULT 'legacy';
