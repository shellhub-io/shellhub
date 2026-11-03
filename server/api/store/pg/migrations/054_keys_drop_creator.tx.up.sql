ALTER TABLE api_keys DROP COLUMN user_id;

--bun:split

ALTER TABLE provisioning_keys DROP COLUMN user_id;
