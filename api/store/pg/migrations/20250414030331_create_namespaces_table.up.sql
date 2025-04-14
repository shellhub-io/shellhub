BEGIN;

DROP TYPE IF EXISTS namespace_scope;
CREATE TYPE namespace_scope AS ENUM ('personal', 'team');

CREATE TABLE IF NOT EXISTS namespaces(
    id UUID PRIMARY KEY,

    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,

    scope namespace_scope NOT NULL,
    name VARCHAR(64) NOT NULL,

    max_devices INTEGER NOT NULL,
    record_sessions BOOLEAN NOT NULL,  
    connection_announcement TEXT
);

COMMIT;
