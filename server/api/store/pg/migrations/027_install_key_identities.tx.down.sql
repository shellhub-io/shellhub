SET LOCAL lock_timeout = '60s';

--bun:split

ALTER TABLE install_keys RENAME COLUMN allowed_identities TO allowed_macs;

--bun:split

ALTER TABLE install_key_events RENAME COLUMN identity TO mac;
