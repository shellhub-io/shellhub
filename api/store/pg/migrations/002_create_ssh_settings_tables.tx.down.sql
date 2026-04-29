-- Down migration: Restore columns and drop new tables

-- Migrate data back from namespace_settings
UPDATE namespaces SET
    record_sessions = ns.record_sessions,
    connection_announcement = ns.connection_announcement
FROM namespace_settings ns
WHERE namespaces.id = ns.namespace_id;

-- Drop tables
DROP TRIGGER IF EXISTS namespace_settings_updated_at ON namespace_settings;
DROP FUNCTION IF EXISTS update_namespace_settings_updated_at();
DROP TABLE IF EXISTS namespace_settings;

DROP TRIGGER IF EXISTS device_settings_updated_at ON device_settings;
DROP FUNCTION IF EXISTS update_device_settings_updated_at();
DROP TABLE IF EXISTS device_settings;
