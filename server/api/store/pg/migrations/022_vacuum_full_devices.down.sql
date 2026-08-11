-- This migration rewrites the devices heap and its indexes; it adds no schema and changes no
-- data, so there is nothing to reverse — reintroducing bloat is not something a rollback
-- should do. Leaving it as a no-op keeps the down runnable.
SELECT 1;
