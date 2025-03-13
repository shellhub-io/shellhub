BEGIN;

CREATE TABLE IF NOT EXISTS devices(
    id BYTEA PRIMARY KEY,

    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,

    last_seen TIMESTAMP WITH TIME ZONE,

    name VARCHAR (50) NOT NULL,

    namespace_id UUID NOT NULL,
    CONSTRAINT fk_namespace FOREIGN KEY (namespace_id) REFERENCES namespaces(id) ON DELETE RESTRICT ON UPDATE RESTRICT
);

CREATE INDEX idx_devices_namespace_id ON devices(namespace_id);

COMMIT;
