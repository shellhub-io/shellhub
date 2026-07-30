-- This migration only backfills registration-history rows; it adds no schema. The synthetic events it
-- writes are indistinguishable from organically recorded ones (append-only, immutable by design), so
-- there is nothing to safely reverse here. Leaving it as a no-op keeps the down runnable without
-- deleting real history.
SELECT 1;
