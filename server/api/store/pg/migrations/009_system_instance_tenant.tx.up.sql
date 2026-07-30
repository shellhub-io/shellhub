-- Bind the ShellHub instance (the `systems` singleton) to its one namespace. In the
-- Community Edition tenant = instance: a set instance_tenant_id makes NamespaceCreate
-- refuse any further namespace, and the FK's ON DELETE RESTRICT protects the bound
-- namespace from deletion. Enterprise/Cloud keep this NULL (the store wrapper strips it
-- on write), leaving multi-tenant unchanged.

-- No data backfill here: a migration can't tell the edition, and setting the binding on
-- an Enterprise instance would break multi-tenant. The API server binds it at boot only
-- when running the core store (Community); see the boot reconciliation step.
-- The constraint is named explicitly so fromSQLError can map its ON DELETE RESTRICT
-- violation (SQLSTATE 23001, or 23503) to store.ErrNamespaceInstanceProtected; renaming
-- it here silently breaks that mapping.
ALTER TABLE systems
    ADD COLUMN IF NOT EXISTS instance_tenant_id uuid NULL
    CONSTRAINT systems_instance_tenant_id_fkey
    REFERENCES namespaces (id) ON DELETE RESTRICT;
