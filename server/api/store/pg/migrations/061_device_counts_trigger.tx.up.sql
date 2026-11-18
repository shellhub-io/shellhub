CREATE FUNCTION namespaces_apply_device_count_deltas(deltas jsonb) RETURNS void
LANGUAGE sql AS $$
    UPDATE namespaces ns SET
        devices_accepted_count = ns.devices_accepted_count + d.accepted,
        devices_pending_count  = ns.devices_pending_count  + d.pending,
        devices_rejected_count = ns.devices_rejected_count + d.rejected,
        devices_removed_count  = ns.devices_removed_count  + d.removed
    FROM (
        SELECT
            (e->>'namespace_id')::uuid AS namespace_id,
            COALESCE(SUM((e->>'n')::bigint) FILTER (WHERE e->>'status' = 'accepted'), 0) AS accepted,
            COALESCE(SUM((e->>'n')::bigint) FILTER (WHERE e->>'status' = 'pending'), 0)  AS pending,
            COALESCE(SUM((e->>'n')::bigint) FILTER (WHERE e->>'status' = 'rejected'), 0) AS rejected,
            COALESCE(SUM((e->>'n')::bigint) FILTER (WHERE e->>'status' = 'removed'), 0)  AS removed
        FROM jsonb_array_elements(deltas) e
        GROUP BY 1
    ) d
    WHERE ns.id = d.namespace_id
      AND (d.accepted, d.pending, d.rejected, d.removed) <> (0, 0, 0, 0);
$$;

CREATE FUNCTION devices_count_on_insert() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    PERFORM namespaces_apply_device_count_deltas(COALESCE((
        SELECT jsonb_agg(jsonb_build_object('namespace_id', namespace_id, 'status', status, 'n', n))
        FROM (SELECT namespace_id, status, count(*) AS n FROM new_rows GROUP BY 1, 2) c
    ), '[]'));

    RETURN NULL;
END;
$$;

CREATE FUNCTION devices_count_on_delete() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    PERFORM namespaces_apply_device_count_deltas(COALESCE((
        SELECT jsonb_agg(jsonb_build_object('namespace_id', namespace_id, 'status', status, 'n', -n))
        FROM (SELECT namespace_id, status, count(*) AS n FROM old_rows GROUP BY 1, 2) c
    ), '[]'));

    RETURN NULL;
END;
$$;

CREATE FUNCTION devices_count_on_update() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    PERFORM namespaces_apply_device_count_deltas(COALESCE((
        SELECT jsonb_agg(jsonb_build_object('namespace_id', namespace_id, 'status', status, 'n', n))
        FROM (
            SELECT namespace_id, status, count(*) AS n FROM new_rows GROUP BY 1, 2
            UNION ALL
            SELECT namespace_id, status, -count(*) AS n FROM old_rows GROUP BY 1, 2
        ) c
    ), '[]'));

    RETURN NULL;
END;
$$;

CREATE TRIGGER devices_count_insert AFTER INSERT ON devices
    REFERENCING NEW TABLE AS new_rows
    FOR EACH STATEMENT EXECUTE FUNCTION devices_count_on_insert();

CREATE TRIGGER devices_count_delete AFTER DELETE ON devices
    REFERENCING OLD TABLE AS old_rows
    FOR EACH STATEMENT EXECUTE FUNCTION devices_count_on_delete();

CREATE TRIGGER devices_count_update AFTER UPDATE ON devices
    REFERENCING OLD TABLE AS old_rows NEW TABLE AS new_rows
    FOR EACH STATEMENT EXECUTE FUNCTION devices_count_on_update();

UPDATE namespaces SET
    devices_accepted_count = (SELECT count(*) FROM devices d WHERE d.namespace_id = namespaces.id AND d.status = 'accepted'),
    devices_pending_count  = (SELECT count(*) FROM devices d WHERE d.namespace_id = namespaces.id AND d.status = 'pending'),
    devices_rejected_count = (SELECT count(*) FROM devices d WHERE d.namespace_id = namespaces.id AND d.status = 'rejected'),
    devices_removed_count  = (SELECT count(*) FROM devices d WHERE d.namespace_id = namespaces.id AND d.status = 'removed');
