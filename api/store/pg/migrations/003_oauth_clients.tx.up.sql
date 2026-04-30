CREATE TABLE oauth_clients (
    id          uuid PRIMARY KEY,
    name        varchar(255) NOT NULL,
    client_id   uuid NOT NULL UNIQUE,
    client_secret varchar NOT NULL,
    namespace_id uuid,
    redirect_uris text[] NOT NULL DEFAULT '{}',
    created_at  timestamptz NOT NULL,
    updated_at  timestamptz NOT NULL
);

--bun:split

CREATE UNIQUE INDEX oauth_clients_namespace_id_name_idx
    ON oauth_clients (namespace_id, name)
    WHERE namespace_id IS NOT NULL;

--bun:split

CREATE INDEX oauth_clients_namespace_id_idx ON oauth_clients (namespace_id);
