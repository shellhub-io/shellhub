CREATE TABLE provisioning_key_tags (
    provisioning_key_digest character(64) NOT NULL,
    provisioning_key_namespace_id uuid NOT NULL,
    tag_id uuid NOT NULL,
    created_at timestamp with time zone NOT NULL,
    PRIMARY KEY (provisioning_key_digest, provisioning_key_namespace_id, tag_id),
    FOREIGN KEY (provisioning_key_digest, provisioning_key_namespace_id) REFERENCES provisioning_keys(key_digest, namespace_id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

--bun:split

CREATE INDEX provisioning_key_tags_tag_id ON provisioning_key_tags USING btree (tag_id);

--bun:split

INSERT INTO tags (id, namespace_id, name, created_at, updated_at)
SELECT gen_random_uuid(), named.namespace_id, named.name, now(), now()
FROM (SELECT DISTINCT namespace_id, unnest(tags) AS name FROM provisioning_keys) AS named
ON CONFLICT (namespace_id, name) DO NOTHING;

--bun:split

INSERT INTO provisioning_key_tags (provisioning_key_digest, provisioning_key_namespace_id, tag_id, created_at)
SELECT DISTINCT k.key_digest, k.namespace_id, t.id, now()
FROM provisioning_keys k
CROSS JOIN LATERAL unnest(k.tags) AS named(name)
JOIN tags t ON t.namespace_id = k.namespace_id AND t.name = named.name;

--bun:split

ALTER TABLE provisioning_keys DROP COLUMN tags;
