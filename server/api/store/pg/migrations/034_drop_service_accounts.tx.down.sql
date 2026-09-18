-- Restores the shape, not the data. The accounts this migration deleted are gone for good.
SET LOCAL lock_timeout = '60s';

--bun:split

CREATE TYPE user_type AS ENUM (
    'human',
    'service'
);

--bun:split

ALTER TABLE users ADD COLUMN type user_type NOT NULL DEFAULT 'human';

--bun:split

ALTER TYPE membership_role ADD VALUE IF NOT EXISTS 'service';
