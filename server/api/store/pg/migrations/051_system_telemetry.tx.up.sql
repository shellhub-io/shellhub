DELETE FROM systems
WHERE id <> (SELECT id FROM systems ORDER BY setup DESC, instance_tenant_id IS NULL, id LIMIT 1);

ALTER TABLE systems
    ADD COLUMN instance_id uuid NOT NULL DEFAULT gen_random_uuid(),
    ADD COLUMN telemetry_enabled boolean NOT NULL DEFAULT true;

CREATE UNIQUE INDEX systems_singleton ON systems ((true));

--bun:split

INSERT INTO systems (id) SELECT gen_random_uuid() WHERE NOT EXISTS (SELECT 1 FROM systems);
