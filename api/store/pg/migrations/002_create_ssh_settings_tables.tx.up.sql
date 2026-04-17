-- Migration 002: Create device_settings table
CREATE TABLE IF NOT EXISTS device_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id character varying NOT NULL UNIQUE REFERENCES devices(id) ON DELETE CASCADE,
    allow_password BOOLEAN DEFAULT TRUE,
    allow_public_key BOOLEAN DEFAULT TRUE,
    allow_root BOOLEAN DEFAULT TRUE,
    allow_empty_passwords BOOLEAN DEFAULT TRUE,
    allow_tty BOOLEAN DEFAULT TRUE,
    allow_tcp_forwarding BOOLEAN DEFAULT TRUE,
    allow_web_endpoints BOOLEAN DEFAULT TRUE,
    allow_sftp BOOLEAN DEFAULT TRUE,
    allow_agent_forwarding BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_device_settings_device_id ON device_settings(device_id);

CREATE OR REPLACE FUNCTION update_device_settings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER device_settings_updated_at
    BEFORE UPDATE ON device_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_device_settings_updated_at();

CREATE TABLE IF NOT EXISTS namespace_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    namespace_id UUID NOT NULL UNIQUE REFERENCES namespaces(id) ON DELETE CASCADE,
    record_sessions BOOLEAN DEFAULT TRUE,
    connection_announcement TEXT DEFAULT '',
    allow_password BOOLEAN DEFAULT TRUE,
    allow_public_key BOOLEAN DEFAULT TRUE,
    allow_root BOOLEAN DEFAULT TRUE,
    allow_empty_passwords BOOLEAN DEFAULT TRUE,
    allow_tty BOOLEAN DEFAULT TRUE,
    allow_tcp_forwarding BOOLEAN DEFAULT TRUE,
    allow_web_endpoints BOOLEAN DEFAULT TRUE,
    allow_sftp BOOLEAN DEFAULT TRUE,
    allow_agent_forwarding BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_namespace_settings_namespace_id ON namespace_settings(namespace_id);

CREATE OR REPLACE FUNCTION update_namespace_settings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER namespace_settings_updated_at
    BEFORE UPDATE ON namespace_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_namespace_settings_updated_at();

-- Migrate data from namespaces table to namespace_settings table
INSERT INTO namespace_settings (namespace_id, record_sessions, connection_announcement, allow_password, allow_public_key, allow_root, allow_empty_passwords, allow_tty, allow_tcp_forwarding, allow_web_endpoints, allow_sftp, allow_agent_forwarding)
SELECT 
    id,
    COALESCE(record_sessions, true),
    COALESCE(connection_announcement, ''),
    TRUE,
    TRUE,
    TRUE,
    TRUE,
    TRUE,
    TRUE,
    TRUE,
    TRUE,
    TRUE
FROM namespaces
ON CONFLICT (namespace_id) DO NOTHING;
