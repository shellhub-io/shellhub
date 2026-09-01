-- WARNING: this migration revokes API keys. Every row in a group sharing one key_digest is
-- deleted, both sides, none elected the survivor. Affected namespaces must re-issue. The digests
-- and namespaces are reported as warnings in the server log before the delete.
DO $$
DECLARE
    collision record;
BEGIN
    FOR collision IN
        SELECT key_digest, count(*) AS shared_by, array_agg(namespace_id) AS namespaces
        FROM api_keys
        GROUP BY key_digest
        HAVING count(*) > 1
    LOOP
        RAISE WARNING 'revoking % API keys sharing digest % across namespaces %',
            collision.shared_by, collision.key_digest, collision.namespaces;
    END LOOP;

    DELETE FROM api_keys
    WHERE key_digest IN (
        SELECT key_digest
        FROM api_keys
        GROUP BY key_digest
        HAVING count(*) > 1
    );
END
$$;

--bun:split

CREATE UNIQUE INDEX api_keys_key_digest_unique ON api_keys USING btree (key_digest);

--bun:split

-- Install keys get the index but no delete. Their plaintexts are server-generated, so a collision
-- is unreachable and this cannot fire; if one exists anyway the migration aborts here rather than
-- deleting, because an install_keys row carries install_key_events by ON DELETE CASCADE and
-- devices by ON DELETE SET NULL, and destroying enrollment history to satisfy a constraint that
-- has never been violated is the wrong trade. An operator who hits this should be told, not
-- silently repaired.
CREATE UNIQUE INDEX install_keys_key_digest_unique ON install_keys USING btree (key_digest);
