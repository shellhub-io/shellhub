ALTER TABLE devices ADD COLUMN key_verified_at timestamptz;

--bun:split

ALTER TABLE systems ADD COLUMN authentication_device_require_verified_key boolean NOT NULL DEFAULT false;

--bun:split

INSERT INTO systems (id)
SELECT gen_random_uuid()
WHERE NOT EXISTS (SELECT 1 FROM systems) AND EXISTS (SELECT 1 FROM namespaces);

--bun:split

ALTER TABLE systems ALTER COLUMN authentication_device_require_verified_key SET DEFAULT true;
