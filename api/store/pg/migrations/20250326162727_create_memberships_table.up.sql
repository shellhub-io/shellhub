BEGIN;

DROP TYPE IF EXISTS membership_status;
CREATE TYPE membership_status AS ENUM ('pending', 'accepted');

DROP TYPE IF EXISTS membership_role;
CREATE TYPE membership_role AS ENUM ('owner', 'administrator', 'operator', 'observer');

CREATE TABLE IF NOT EXISTS memberships(
    user_id UUID NOT NULL,
    namespace_id UUID NOT NULL,

    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,

    status membership_status NOT NULL,
    role membership_role NOT NULL,

    PRIMARY KEY (user_id, namespace_id),
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_namespace FOREIGN KEY (namespace_id) REFERENCES namespaces(id)
);

CREATE INDEX idx_memberships_user_id ON memberships(user_id);
CREATE INDEX idx_memberships_namespace_id ON memberships(namespace_id);

COMMIT;
