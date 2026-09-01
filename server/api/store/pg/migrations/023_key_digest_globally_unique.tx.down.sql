-- Drops the indexes only. The API keys the up migration revoked are not recoverable.
DROP INDEX IF EXISTS install_keys_key_digest_unique;

--bun:split

DROP INDEX IF EXISTS api_keys_key_digest_unique;
