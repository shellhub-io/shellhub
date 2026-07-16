-- Restore the namespace-wide auto-accept setting this migration dropped.
ALTER TABLE namespaces ADD COLUMN IF NOT EXISTS device_auto_accept boolean DEFAULT false NOT NULL;

--bun:split

DROP TABLE IF EXISTS enrollment_callback_redemptions;

--bun:split

-- install_key_events references install_keys, so it goes first.
DROP TABLE IF EXISTS install_key_events;

--bun:split

-- Drop the device columns before install_keys: install_key_id carries a FK to it.
ALTER TABLE devices DROP COLUMN IF EXISTS install_key_id;

--bun:split

ALTER TABLE devices DROP COLUMN IF EXISTS ephemeral;

--bun:split

ALTER TABLE devices DROP COLUMN IF EXISTS ephemeral_timeout;

--bun:split

ALTER TABLE devices DROP COLUMN IF EXISTS last_enrollment_attempt_at;

--bun:split

DROP TABLE IF EXISTS install_keys;
