SELECT count(*) FROM (SELECT 1 FROM namespaces ORDER BY id FOR NO KEY UPDATE) locked;

--bun:split

UPDATE namespaces ns SET
    devices_accepted_count = c.accepted,
    devices_pending_count  = c.pending,
    devices_rejected_count = c.rejected,
    devices_removed_count  = c.removed
FROM (
    SELECT
        n.id,
        count(*) FILTER (WHERE d.status = 'accepted') AS accepted,
        count(*) FILTER (WHERE d.status = 'pending')  AS pending,
        count(*) FILTER (WHERE d.status = 'rejected') AS rejected,
        count(*) FILTER (WHERE d.status = 'removed')  AS removed
    FROM namespaces n
    LEFT JOIN devices d ON d.namespace_id = n.id
    GROUP BY n.id
) c
WHERE ns.id = c.id
  AND (ns.devices_accepted_count, ns.devices_pending_count, ns.devices_rejected_count, ns.devices_removed_count)
      IS DISTINCT FROM (c.accepted, c.pending, c.rejected, c.removed);
