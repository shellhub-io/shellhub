BEGIN;

DROP TYPE IF EXISTS device_status;
CREATE TYPE device_status AS ENUM ('accepted', 'pending', 'rejected', 'removed', 'unused');

CREATE TABLE IF NOT EXISTS devices (
    id VARCHAR PRIMARY KEY,
    namespace_id UUID NOT NULL,
    
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    seen_at TIMESTAMPTZ NOT NULL,
    disconnected_at TIMESTAMPTZ,
    
    status device_status NOT NULL,
    name VARCHAR(64) NOT NULL,
    mac VARCHAR(17) NOT NULL,
    public_key TEXT NOT NULL,
    
    CONSTRAINT fk_namespace FOREIGN KEY (namespace_id) REFERENCES namespaces(id)
);

CREATE INDEX idx_devices_namespace_id ON devices(namespace_id);
CREATE INDEX idx_devices_seen_at ON devices(seen_at);
CREATE INDEX idx_devices_disconnected_at ON devices(disconnected_at);

CREATE TABLE IF NOT EXISTS device_info (
    device_id VARCHAR PRIMARY KEY,
    
    identifier VARCHAR NOT NULL,
    pretty_name VARCHAR(64) NOT NULL,
    version VARCHAR(32) NOT NULL,
    arch VARCHAR(16) NOT NULL,
    platform VARCHAR(32) NOT NULL,
    
    CONSTRAINT fk_device FOREIGN KEY (device_id) REFERENCES devices(id)
);

CREATE TABLE IF NOT EXISTS device_position (
    device_id VARCHAR PRIMARY KEY,
    
    latitude NUMERIC,
    longitude NUMERIC,
    
    CONSTRAINT fk_device FOREIGN KEY (device_id) REFERENCES devices(id)
);

COMMIT;
