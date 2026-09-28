ALTER TABLE devices ADD COLUMN owner_id uuid;

--bun:split

ALTER TABLE devices
    ADD CONSTRAINT devices_owner_member_fkey
    FOREIGN KEY (owner_id, namespace_id)
    REFERENCES memberships (user_id, namespace_id)
    DEFERRABLE INITIALLY DEFERRED;

--bun:split

CREATE INDEX devices_namespace_owner_idx ON devices (namespace_id, owner_id) WHERE owner_id IS NOT NULL;
