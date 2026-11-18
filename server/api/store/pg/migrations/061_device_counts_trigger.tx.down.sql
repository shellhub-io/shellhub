DROP TRIGGER devices_count_update ON devices;
DROP TRIGGER devices_count_delete ON devices;
DROP TRIGGER devices_count_insert ON devices;
DROP FUNCTION devices_count_on_update();
DROP FUNCTION devices_count_on_delete();
DROP FUNCTION devices_count_on_insert();
DROP FUNCTION namespaces_apply_device_count_deltas(jsonb);
