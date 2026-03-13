CREATE TYPE device_status AS ENUM (
    'accepted',
    'pending',
    'rejected',
    'removed',
    'unused'
);

--bun:split

CREATE TYPE membership_role AS ENUM (
    'owner',
    'administrator',
    'operator',
    'observer'
);

--bun:split

CREATE TYPE namespace_scope AS ENUM (
    'personal',
    'team'
);

--bun:split

CREATE TYPE session_type AS ENUM (
    'shell',
    'exec',
    'scp',
    'sftp',
    'subsystem',
    'term',
    'web',
    'heredoc',
    'unknown',
    'none'
);

--bun:split

CREATE TYPE user_auth_method AS ENUM (
    'local',
    'saml'
);

--bun:split

CREATE TYPE user_origin AS ENUM (
    'local',
    'saml'
);

--bun:split

CREATE TYPE user_status AS ENUM (
    'not-confirmed',
    'confirmed'
);

--bun:split

CREATE TABLE namespaces (
    id uuid NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    scope namespace_scope NOT NULL,
    name character varying(64) NOT NULL,
    owner_id uuid NOT NULL,
    max_devices integer NOT NULL,
    record_sessions boolean NOT NULL,
    connection_announcement text,
    devices_accepted_count bigint DEFAULT 0 NOT NULL,
    devices_pending_count bigint DEFAULT 0 NOT NULL,
    devices_rejected_count bigint DEFAULT 0 NOT NULL,
    devices_removed_count bigint DEFAULT 0 NOT NULL,
    PRIMARY KEY (id)
);

--bun:split

CREATE TABLE users (
    id uuid NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    last_login timestamp with time zone,
    origin user_origin NOT NULL,
    external_id character varying,
    status user_status NOT NULL,
    name character varying(64) NOT NULL,
    username character varying(64) NOT NULL,
    email character varying(320) NOT NULL,
    security_email character varying(320),
    password_digest character varying NOT NULL,
    auth_methods user_auth_method[] NOT NULL,
    namespace_ownership_limit integer NOT NULL,
    email_marketing boolean DEFAULT false NOT NULL,
    preferred_namespace_id uuid,
    admin boolean DEFAULT false NOT NULL,
    PRIMARY KEY (id),
    UNIQUE (email),
    UNIQUE (username),
    FOREIGN KEY (preferred_namespace_id) REFERENCES namespaces(id) ON DELETE SET NULL
);

--bun:split

ALTER TABLE namespaces ADD FOREIGN KEY (owner_id)
    REFERENCES users(id) ON DELETE CASCADE
    DEFERRABLE INITIALLY DEFERRED;

--bun:split

CREATE TABLE memberships (
    user_id uuid NOT NULL,
    namespace_id uuid NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    role membership_role NOT NULL,
    PRIMARY KEY (user_id, namespace_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (namespace_id) REFERENCES namespaces(id) ON DELETE CASCADE
);

--bun:split

CREATE TABLE devices (
    id character varying NOT NULL,
    namespace_id uuid NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    removed_at timestamp with time zone,
    last_seen timestamp with time zone NOT NULL,
    disconnected_at timestamp with time zone,
    status device_status NOT NULL,
    name character varying(64) NOT NULL,
    mac character varying(64) NOT NULL,
    public_key text NOT NULL,
    identifier character varying,
    pretty_name character varying(64),
    version character varying(32),
    arch character varying(16),
    platform character varying(32),
    latitude numeric,
    longitude numeric,
    status_updated_at timestamp with time zone,
    PRIMARY KEY (id),
    FOREIGN KEY (namespace_id) REFERENCES namespaces(id) ON DELETE CASCADE
);

--bun:split

CREATE INDEX devices_namespace_id ON devices USING btree (namespace_id);

--bun:split

CREATE INDEX devices_last_seen ON devices USING btree (last_seen);

--bun:split

CREATE INDEX devices_disconnected_at ON devices USING btree (disconnected_at);

--bun:split

CREATE TABLE private_keys (
    fingerprint character varying NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    data bytea,
    PRIMARY KEY (fingerprint)
);

--bun:split

CREATE TABLE api_keys (
    key_digest character(64) NOT NULL,
    namespace_id uuid NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    expires_in bigint,
    name character varying NOT NULL,
    role membership_role NOT NULL,
    user_id uuid NOT NULL,
    PRIMARY KEY (key_digest, namespace_id),
    FOREIGN KEY (namespace_id) REFERENCES namespaces(id) ON DELETE CASCADE
);

--bun:split

CREATE UNIQUE INDEX api_keys_namespace_id_name_unique ON api_keys USING btree (namespace_id, name);

--bun:split

CREATE TABLE public_keys (
    fingerprint character(47) NOT NULL,
    namespace_id uuid NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    name character varying NOT NULL,
    username character varying DEFAULT ''::character varying NOT NULL,
    data bytea,
    filter_hostname character varying DEFAULT ''::character varying,
    PRIMARY KEY (fingerprint, namespace_id),
    FOREIGN KEY (namespace_id) REFERENCES namespaces(id) ON DELETE CASCADE
);

--bun:split

CREATE TABLE tags (
    id uuid NOT NULL,
    namespace_id uuid NOT NULL,
    name character varying NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (namespace_id) REFERENCES namespaces(id) ON DELETE CASCADE
);

--bun:split

CREATE UNIQUE INDEX tags_namespace_id_name_unique ON tags USING btree (namespace_id, name);

--bun:split

CREATE TABLE device_tags (
    device_id character varying NOT NULL,
    tag_id uuid NOT NULL,
    created_at timestamp with time zone NOT NULL,
    PRIMARY KEY (device_id, tag_id),
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

--bun:split

CREATE INDEX device_tags_device_id ON device_tags USING btree (device_id);

--bun:split

CREATE INDEX device_tags_tag_id ON device_tags USING btree (tag_id);

--bun:split

CREATE TABLE public_key_tags (
    public_key_fingerprint character(47) NOT NULL,
    public_key_namespace_id uuid NOT NULL,
    tag_id uuid NOT NULL,
    created_at timestamp with time zone NOT NULL,
    PRIMARY KEY (public_key_fingerprint, public_key_namespace_id, tag_id),
    FOREIGN KEY (public_key_fingerprint, public_key_namespace_id) REFERENCES public_keys(fingerprint, namespace_id) ON DELETE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES tags(id) ON DELETE CASCADE
);

--bun:split

CREATE INDEX public_key_tags_public_key_fingerprint ON public_key_tags USING btree (public_key_fingerprint);

--bun:split

CREATE INDEX public_key_tags_tag_id ON public_key_tags USING btree (tag_id);

--bun:split

CREATE TABLE systems (
    id uuid NOT NULL,
    setup boolean DEFAULT false NOT NULL,
    authentication_local_enabled boolean DEFAULT true NOT NULL,
    authentication_saml_enabled boolean DEFAULT false NOT NULL,
    authentication_saml_idp_entity_id text,
    authentication_saml_idp_binding_post text,
    authentication_saml_idp_binding_redirect text,
    authentication_saml_idp_binding_preferred text,
    authentication_saml_idp_certificates character varying[],
    authentication_saml_idp_mappings jsonb,
    authentication_saml_sp_sign_auth_requests boolean DEFAULT false NOT NULL,
    authentication_saml_sp_certificate text,
    authentication_saml_sp_private_key text,
    PRIMARY KEY (id)
);

--bun:split

CREATE TABLE sessions (
    id character varying(128) NOT NULL,
    device_id character varying NOT NULL,
    username character varying(64) NOT NULL,
    ip_address inet NOT NULL,
    started_at timestamp with time zone NOT NULL,
    seen_at timestamp with time zone NOT NULL,
    closed boolean DEFAULT false NOT NULL,
    authenticated boolean DEFAULT false NOT NULL,
    recorded boolean DEFAULT false NOT NULL,
    type session_type,
    term character varying(32),
    longitude numeric(10,7),
    latitude numeric(10,7),
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (device_id) REFERENCES devices(id) ON DELETE CASCADE
);

--bun:split

CREATE INDEX sessions_device_id_idx ON sessions USING btree (device_id);

--bun:split

CREATE INDEX sessions_started_at_idx ON sessions USING btree (started_at);

--bun:split

CREATE INDEX sessions_closed_started_idx ON sessions USING btree (closed, started_at);

--bun:split

CREATE INDEX sessions_username_idx ON sessions USING btree (username);

--bun:split

CREATE INDEX sessions_type_idx ON sessions USING btree (type);

--bun:split

CREATE TABLE active_sessions (
    session_id character varying(128) NOT NULL,
    seen_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    PRIMARY KEY (session_id),
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

--bun:split

CREATE TABLE session_events (
    id uuid NOT NULL,
    session_id character varying(128) NOT NULL,
    type character varying(64) NOT NULL,
    seat integer NOT NULL,
    data text,
    created_at timestamp with time zone NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

--bun:split

CREATE INDEX session_events_seat_idx ON session_events USING btree (seat);

--bun:split

CREATE INDEX session_events_session_id_created_at_idx ON session_events USING btree (session_id, created_at);

--bun:split

CREATE INDEX session_events_type_created_at_idx ON session_events USING btree (type, created_at);


