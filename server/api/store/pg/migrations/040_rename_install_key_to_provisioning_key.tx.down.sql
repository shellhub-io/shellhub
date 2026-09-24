ALTER TABLE provisioning_keys RENAME TO install_keys;

--bun:split

ALTER TABLE provisioning_key_events RENAME TO install_key_events;

--bun:split

ALTER TABLE install_key_events RENAME COLUMN provisioning_key_id TO install_key_id;

--bun:split

ALTER TABLE devices RENAME COLUMN provisioning_key_id TO install_key_id;

--bun:split

DO $$
DECLARE
    renamed record;
BEGIN
    FOR renamed IN
        SELECT conrelid::regclass AS relation, conname
        FROM pg_constraint
        WHERE connamespace = 'public'::regnamespace AND conname LIKE '%provisioning_key%'
    LOOP
        EXECUTE format('ALTER TABLE %s RENAME CONSTRAINT %I TO %I',
            renamed.relation, renamed.conname, replace(renamed.conname, 'provisioning_key', 'install_key'));
    END LOOP;

    FOR renamed IN
        SELECT indexname
        FROM pg_indexes
        WHERE schemaname = 'public' AND indexname LIKE '%provisioning_key%'
    LOOP
        EXECUTE format('ALTER INDEX %I RENAME TO %I',
            renamed.indexname, replace(renamed.indexname, 'provisioning_key', 'install_key'));
    END LOOP;
END
$$;
