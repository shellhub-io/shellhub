CREATE TABLE IF NOT EXISTS private_keys (
    fingerprint character varying NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    data bytea,
    PRIMARY KEY (fingerprint)
);
