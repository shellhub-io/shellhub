-- Three diagnostic extensions the image already ships, none of which ShellHub itself queries:
-- pg_stat_statements attributes load to statements rather than tables, pgstattuple measures
-- bloat and in-page free space directly, and pg_buffercache shows what occupies shared_buffers.
-- Without the first, a scan storm can be traced to a table but never to the query behind it.
--
-- Each CREATE is wrapped so it can only warn, never fail. Migrations run inline in Server.Setup
-- before the listener binds and an error there is fatal, so an unguarded CREATE EXTENSION would
-- turn a database role without superuser -- an externally managed Postgres, say -- into a server
-- that cannot boot. None of these three is a trusted extension, so that role is a real one.
-- WHEN OTHERS rather than insufficient_privilege alone also covers a build without contrib,
-- where the control file is simply absent.
--
-- The trade is that a swallowed failure is not retried: bun marks a migration applied before
-- running it. That is deliberate for diagnostics, and CREATE EXTENSION IF NOT EXISTS means an
-- operator can finish the job by hand at any time.

-- The only one of the three that also needs shared_preload_libraries, set on the postgres
-- command line. Creating it without that succeeds -- the install script only defines functions
-- and a view, and the library loads lazily -- so the extension is safe to create anywhere and
-- merely reports an error on read until the server is restarted with the library preloaded.
DO $$ BEGIN
    CREATE EXTENSION IF NOT EXISTS pg_stat_statements;
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'skipping pg_stat_statements: %', SQLERRM;
END $$;

--bun:split

DO $$ BEGIN
    CREATE EXTENSION IF NOT EXISTS pgstattuple;
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'skipping pgstattuple: %', SQLERRM;
END $$;

--bun:split

DO $$ BEGIN
    CREATE EXTENSION IF NOT EXISTS pg_buffercache;
EXCEPTION WHEN OTHERS THEN
    RAISE NOTICE 'skipping pg_buffercache: %', SQLERRM;
END $$;
