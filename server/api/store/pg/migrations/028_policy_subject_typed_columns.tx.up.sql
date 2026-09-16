-- An access policy names its subject in one polymorphic column: subject_value held a user id, a
-- role name, or nothing, depending on subject_type. Nothing could constrain it, so the database
-- accepted a user who was never a member, a role that does not exist, and a value on a subject
-- that matches by type alone. Each of those stores a rule that can never match anyone: an allow
-- that grants nothing, or a deny that blocks nothing while the console shows it as in force.
--
-- Splitting it into two typed columns moves those checks into the schema. The composite foreign
-- key to memberships also decides what happens when the member leaves: the policy naming them goes
-- with the membership, rather than lingering as a rule nobody can match.
--
-- ssh_identities gets the same foreign key for the same reason. Its user_id already cascades from
-- users, so deleting an account revoked the keys, but removing someone from a namespace left them
-- enrolled there, ready to work again if the person was ever re-added.
--
-- lock_timeout because migrations run inline at startup, and every statement here takes ACCESS
-- EXCLUSIVE on a table the API is about to read.
SET LOCAL lock_timeout = '60s';

--bun:split

-- The comparison casts the column to text rather than the value to uuid: subject_value is free
-- text that nothing constrained, so it can hold something no uuid cast would survive, and a row
-- like that is not a member either. Comparing as text lets one statement drop both cases, and
-- leaves only values the backfill below can cast.
DELETE FROM access_policies
WHERE subject_type = 'user'
  AND NOT EXISTS (
    SELECT 1 FROM memberships m
    WHERE m.user_id::text = subject_value AND m.namespace_id = access_policies.namespace_id
  );

--bun:split

DELETE FROM access_policies
WHERE subject_type = 'role'
  AND subject_value NOT IN (SELECT unnest(enum_range(NULL::membership_role))::text);

--bun:split

DELETE FROM ssh_identities
WHERE NOT EXISTS (
  SELECT 1 FROM memberships m
  WHERE m.user_id = ssh_identities.user_id AND m.namespace_id = ssh_identities.namespace_id
);

--bun:split

ALTER TABLE access_policies
    ADD COLUMN subject_user_id uuid,
    ADD COLUMN subject_role membership_role;

--bun:split

UPDATE access_policies SET subject_user_id = subject_value::uuid WHERE subject_type = 'user';

--bun:split

UPDATE access_policies SET subject_role = subject_value::membership_role WHERE subject_type = 'role';

--bun:split

ALTER TABLE access_policies DROP COLUMN subject_value;

--bun:split

ALTER TABLE access_policies
    ADD CONSTRAINT access_policies_subject_shape CHECK (
        (subject_type = 'user' AND subject_user_id IS NOT NULL AND subject_role IS NULL)
        OR (subject_type = 'role' AND subject_role IS NOT NULL AND subject_user_id IS NULL)
        OR (subject_type = 'all-members' AND subject_user_id IS NULL AND subject_role IS NULL)
    );

--bun:split

ALTER TABLE access_policies
    ADD CONSTRAINT access_policies_subject_member_fkey
    FOREIGN KEY (subject_user_id, namespace_id)
    REFERENCES memberships (user_id, namespace_id) ON DELETE CASCADE;

--bun:split

ALTER TABLE ssh_identities
    ADD CONSTRAINT ssh_identities_member_fkey
    FOREIGN KEY (user_id, namespace_id)
    REFERENCES memberships (user_id, namespace_id) ON DELETE CASCADE;
