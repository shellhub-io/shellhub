BEGIN;

DROP TYPE IF EXISTS user_origin;
CREATE TYPE user_origin AS ENUM ('local', 'saml');

DROP TYPE IF EXISTS user_status;
CREATE TYPE user_status AS ENUM ('invited', 'pending', 'confirmed');

DROP TYPE IF EXISTS user_auth_method;
CREATE TYPE user_auth_method AS ENUM ('local', 'saml');


CREATE TABLE IF NOT EXISTS users(
    id UUID PRIMARY KEY,

    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    last_login TIMESTAMPTZ,

    origin user_origin NOT NULL,
    external_id VARCHAR,
    status user_status NOT NULL,
    name VARCHAR(64) NOT NULL,
    username VARCHAR(32) UNIQUE NOT NULL,
    email VARCHAR(320) UNIQUE NOT NULL,
    security_email VARCHAR(320),
    password_digest CHAR(72) NOT NULL,
    auth_methods user_auth_method[] NOT NULL,

    namespace_ownership_limit INTEGER NOT NULL,
    email_marketing BOOLEAN NOT NULL,  
    preferred_namespace_id UUID
);

COMMIT;
