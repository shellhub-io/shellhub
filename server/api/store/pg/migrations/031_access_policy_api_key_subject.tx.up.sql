-- An access policy can now name an API key, which is how an automation is told where it may
-- connect. It is a typed column beside the other two rather than a value in a shared one, for
-- the reason 028 gives: a value nothing constrains stores a rule that can never match anyone.
--
-- The cascade mirrors what 028 did for the member subject. Revoking the key removes the policy
-- naming it, instead of leaving a rule the console still renders as in force.
SET LOCAL lock_timeout = '60s';

--bun:split

ALTER TABLE access_policies
    ADD COLUMN subject_api_key_id uuid;

--bun:split

ALTER TABLE access_policies
    DROP CONSTRAINT access_policies_subject_shape;

--bun:split

ALTER TABLE access_policies
    ADD CONSTRAINT access_policies_subject_shape CHECK (
        (subject_type = 'user' AND subject_user_id IS NOT NULL AND subject_role IS NULL AND subject_api_key_id IS NULL)
        OR (subject_type = 'role' AND subject_role IS NOT NULL AND subject_user_id IS NULL AND subject_api_key_id IS NULL)
        OR (subject_type = 'api-key' AND subject_api_key_id IS NOT NULL AND subject_user_id IS NULL AND subject_role IS NULL)
        OR (subject_type = 'all-members' AND subject_user_id IS NULL AND subject_role IS NULL AND subject_api_key_id IS NULL)
    );

--bun:split

ALTER TABLE access_policies
    ADD CONSTRAINT access_policies_subject_api_key_fkey
    FOREIGN KEY (subject_api_key_id, namespace_id)
    REFERENCES api_keys (id, namespace_id) ON DELETE CASCADE;
