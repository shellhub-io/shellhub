ALTER TABLE api_keys ADD COLUMN user_id uuid;

--bun:split

ALTER TABLE provisioning_keys ADD COLUMN user_id uuid;
