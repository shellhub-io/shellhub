ALTER TABLE access_policies ADD COLUMN all_devices boolean NOT NULL DEFAULT false;

--bun:split

UPDATE access_policies
SET all_devices = true
WHERE NOT EXISTS (
    SELECT 1 FROM access_policy_tags WHERE access_policy_tags.access_policy_id = access_policies.id
);
