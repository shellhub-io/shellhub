CREATE TABLE connections (
    id uuid NOT NULL,
    namespace_id uuid NOT NULL,
    label character varying NOT NULL,
    username character varying NOT NULL DEFAULT '',
    kind character varying NOT NULL DEFAULT 'direct',
    host character varying NOT NULL DEFAULT '',
    port integer NOT NULL DEFAULT 22,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (namespace_id) REFERENCES namespaces(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX connections_namespace_id_label_unique ON connections USING btree (namespace_id, label);
