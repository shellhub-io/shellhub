CREATE TABLE issued_tenants (
    id uuid NOT NULL,
    created_at timestamp with time zone NOT NULL,
    PRIMARY KEY (id)
);

--bun:split

INSERT INTO issued_tenants (id, created_at)
SELECT id, created_at FROM namespaces;

--bun:split

CREATE FUNCTION issued_tenants_append_only() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
    RAISE EXCEPTION 'issued_tenants is append-only' USING ERRCODE = 'restrict_violation';
END;
$$;

--bun:split

CREATE TRIGGER issued_tenants_append_only
    BEFORE UPDATE OR DELETE ON issued_tenants
    FOR EACH ROW EXECUTE FUNCTION issued_tenants_append_only();

--bun:split

ALTER TABLE namespaces
    ADD CONSTRAINT namespaces_issued_tenant_fkey FOREIGN KEY (id) REFERENCES issued_tenants (id);
