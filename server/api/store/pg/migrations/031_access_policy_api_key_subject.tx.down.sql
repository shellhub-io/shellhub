-- Reverting drops every policy naming an API key: the subject type is about to be uncheckable
-- and those rows would fail the restored constraint.
SET LOCAL lock_timeout = '60s';

--bun:split

DELETE FROM access_policies WHERE subject_type = 'api-key';

--bun:split

ALTER TABLE access_policies
    DROP CONSTRAINT access_policies_subject_api_key_fkey;

--bun:split

ALTER TABLE access_policies
    DROP CONSTRAINT access_policies_subject_shape;

--bun:split

ALTER TABLE access_policies
    ADD CONSTRAINT access_policies_subject_shape CHECK (
        (subject_type = 'user' AND subject_user_id IS NOT NULL AND subject_role IS NULL)
        OR (subject_type = 'role' AND subject_role IS NOT NULL AND subject_user_id IS NULL)
        OR (subject_type = 'all-members' AND subject_user_id IS NULL AND subject_role IS NULL)
    );

--bun:split

ALTER TABLE access_policies
    DROP COLUMN subject_api_key_id;
