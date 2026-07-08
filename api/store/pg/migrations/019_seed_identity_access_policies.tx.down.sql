-- No-op: the seeded starter policies are indistinguishable from user-authored ones
-- once created (a user may rename or keep "Default access"), so we do not delete
-- them on rollback to avoid removing real grants.
SELECT 1;
