-- Reverse the type discriminator: restore the `system` boolean, drop the pairing keys, drop `type`.
-- The devices FK is ON DELETE SET NULL, so any device attributed to a pairing key has its
-- install_key_id cleared rather than blocking the drop.
DROP INDEX IF EXISTS install_keys_namespace_type_unique;

ALTER TABLE install_keys ADD COLUMN system boolean NOT NULL DEFAULT false;
UPDATE install_keys SET system = true WHERE type <> 'user';

DELETE FROM install_keys WHERE type = 'pairing';

ALTER TABLE install_keys DROP COLUMN type;
