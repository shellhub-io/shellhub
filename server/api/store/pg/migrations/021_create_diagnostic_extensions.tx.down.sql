-- Guarded for the same reason as the up migration: a role that could not create these cannot
-- drop them either, and neither direction is worth failing a boot over.
DO $$ BEGIN
    DROP EXTENSION IF EXISTS pg_stat_statements;
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'skipping pg_stat_statements: %', SQLERRM;
END $$;

--bun:split

DO $$ BEGIN
    DROP EXTENSION IF EXISTS pgstattuple;
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'skipping pgstattuple: %', SQLERRM;
END $$;

--bun:split

DO $$ BEGIN
    DROP EXTENSION IF EXISTS pg_buffercache;
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'skipping pg_buffercache: %', SQLERRM;
END $$;
