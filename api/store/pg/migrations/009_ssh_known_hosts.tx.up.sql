CREATE TABLE ssh_known_hosts (
    id uuid NOT NULL,
    namespace_id uuid NOT NULL,
    owner_id uuid,
    host character varying NOT NULL,
    port integer NOT NULL,
    key_type character varying NOT NULL,
    public_key text NOT NULL,
    fingerprint character varying NOT NULL,
    accepted_by uuid,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (namespace_id) REFERENCES namespaces(id) ON DELETE CASCADE,
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (accepted_by) REFERENCES users(id) ON DELETE SET NULL
);

-- A known host follows the scope of the connection it was reached through:
-- personal (owner_id set) is per-user; team (owner_id NULL) is shared per namespace.
CREATE UNIQUE INDEX ssh_known_hosts_personal_unique ON ssh_known_hosts USING btree (namespace_id, owner_id, host, port) WHERE owner_id IS NOT NULL;
CREATE UNIQUE INDEX ssh_known_hosts_team_unique ON ssh_known_hosts USING btree (namespace_id, host, port) WHERE owner_id IS NULL;
