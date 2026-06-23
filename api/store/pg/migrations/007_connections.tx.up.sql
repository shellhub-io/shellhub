CREATE TABLE connections (
    id uuid NOT NULL,
    namespace_id uuid NOT NULL,
    owner_id uuid NOT NULL,
    label character varying NOT NULL,
    kind character varying NOT NULL DEFAULT 'external',
    host character varying NOT NULL DEFAULT '',
    port integer NOT NULL DEFAULT 22,
    username character varying NOT NULL DEFAULT '',
    auth_method character varying NOT NULL DEFAULT '',
    key_fingerprint character varying NOT NULL DEFAULT '',
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (namespace_id) REFERENCES namespaces(id) ON DELETE CASCADE,
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);

-- A connection is personal; labels are unique per owner within a namespace.
CREATE UNIQUE INDEX connections_ns_owner_label_unique ON connections USING btree (namespace_id, owner_id, label);
