ALTER TABLE systems DROP COLUMN authentication_device_require_verified_key;

--bun:split

ALTER TABLE devices DROP COLUMN key_verified_at;
