-- Only the constraints come off. The up migration's backfill granted the legacy permission to rows
-- that were in legacy mode without it, and that grant cannot be told apart from one a grandfathered
-- namespace has legitimately carried since 016. Reversing it by predicate would strip the
-- permission from namespaces that always had it, so the data stays as it is: the values it
-- replaced were the defect.
ALTER TABLE namespaces
    DROP CONSTRAINT namespaces_legacy_requires_permission,
    DROP CONSTRAINT namespaces_ssh_access_mode_check;
