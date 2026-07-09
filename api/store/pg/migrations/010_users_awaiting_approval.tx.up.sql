-- Track accounts a namespace admin provisioned but a system admin has not approved yet.
-- While awaiting_approval is true the account is inert: it exists as not-confirmed and only
-- an admin can mint its activation link. It is set false when an admin creates the account
-- directly (auto-approved) or approves a pending one. Community never sets it true (non-admin
-- provisioning is an enterprise capability), so the default false keeps existing rows correct.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS awaiting_approval boolean NOT NULL DEFAULT false;
