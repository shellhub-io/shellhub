WITH newest AS (
    SELECT DISTINCT ON (tenant_id, user_id) id
    FROM membership_invitations
    ORDER BY tenant_id, user_id, created_at DESC, id DESC
)
UPDATE membership_invitations
SET status = 'cancelled',
    status_updated_at = now(),
    updated_at = now()
WHERE status = 'pending'
  AND id NOT IN (SELECT id FROM newest);

--bun:split

CREATE UNIQUE INDEX membership_invitations_pending_unique
    ON membership_invitations USING btree (tenant_id, user_id)
    WHERE status = 'pending';
