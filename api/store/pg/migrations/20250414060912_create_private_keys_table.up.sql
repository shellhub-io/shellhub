BEGIN;

CREATE TABLE IF NOT EXISTS private_keys(
    fingerprint VARCHAR PRIMARY KEY,

    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,

    data BYTEA
);

COMMIT;
