SET LOCAL lock_timeout = '60s';

--bun:split

ALTER TABLE ssh_identities DROP CONSTRAINT IF EXISTS ssh_identities_member_fkey;

--bun:split

ALTER TABLE access_policies DROP CONSTRAINT IF EXISTS access_policies_subject_member_fkey;

--bun:split

ALTER TABLE access_policies DROP CONSTRAINT IF EXISTS access_policies_subject_shape;

--bun:split

ALTER TABLE access_policies ADD COLUMN subject_value character varying NOT NULL DEFAULT '';

--bun:split

UPDATE access_policies SET subject_value = subject_user_id::text WHERE subject_user_id IS NOT NULL;

--bun:split

UPDATE access_policies SET subject_value = subject_role::text WHERE subject_role IS NOT NULL;

--bun:split

ALTER TABLE access_policies DROP COLUMN subject_user_id, DROP COLUMN subject_role;
