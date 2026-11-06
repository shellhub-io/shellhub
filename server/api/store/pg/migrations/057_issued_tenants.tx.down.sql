ALTER TABLE namespaces DROP CONSTRAINT IF EXISTS namespaces_issued_tenant_fkey;

--bun:split

DROP TABLE IF EXISTS issued_tenants;

--bun:split

DROP FUNCTION IF EXISTS issued_tenants_append_only();
