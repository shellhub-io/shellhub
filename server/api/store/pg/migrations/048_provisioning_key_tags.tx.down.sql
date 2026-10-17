ALTER TABLE provisioning_keys ADD COLUMN tags text[] DEFAULT '{}' NOT NULL;

--bun:split

UPDATE provisioning_keys k
SET tags = linked.names
FROM (
    SELECT kt.provisioning_key_digest, kt.provisioning_key_namespace_id, array_agg(t.name ORDER BY kt.created_at, t.name) AS names
    FROM provisioning_key_tags kt
    JOIN tags t ON t.id = kt.tag_id
    GROUP BY kt.provisioning_key_digest, kt.provisioning_key_namespace_id
) AS linked
WHERE k.key_digest = linked.provisioning_key_digest AND k.namespace_id = linked.provisioning_key_namespace_id;

--bun:split

DROP TABLE provisioning_key_tags;
