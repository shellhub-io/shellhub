-- SSH Identities: bindings of an SSH public key to a ShellHub account within a
-- namespace, for the identity SSH access mode. A connection whose presented
-- key's fingerprint resolves here is recognized as the bound account, with no
-- browser step. Distinct from public_keys (the legacy namespace-scoped key ACL),
-- which is untouched.
CREATE TABLE ssh_identities (
    id uuid NOT NULL,
    namespace_id uuid NOT NULL,
    user_id uuid NOT NULL,
    -- SSH public key fingerprint, "SHA256:…" form (50 chars); varchar to fit it.
    fingerprint character varying NOT NULL,
    data bytea NOT NULL,
    name character varying NOT NULL DEFAULT '',
    created_at timestamp with time zone NOT NULL,
    last_used_at timestamp with time zone,
    PRIMARY KEY (id),
    -- A fingerprint maps to exactly one identity within a namespace; the same key
    -- may still enroll in other namespaces.
    UNIQUE (namespace_id, fingerprint),
    FOREIGN KEY (namespace_id) REFERENCES namespaces(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

--bun:split

CREATE INDEX ssh_identities_namespace_id ON ssh_identities USING btree (namespace_id);

--bun:split

CREATE INDEX ssh_identities_user_id ON ssh_identities USING btree (user_id);
