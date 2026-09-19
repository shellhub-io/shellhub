ALTER TABLE ssh_approvals
    ADD COLUMN confirmation_code VARCHAR NOT NULL DEFAULT '';
