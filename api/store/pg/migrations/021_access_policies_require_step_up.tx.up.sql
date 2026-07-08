-- Per-policy JIT step-up: access granted by a policy with this flag set still
-- triggers a per-session browser approval, even for an already-enrolled key.
ALTER TABLE access_policies ADD COLUMN IF NOT EXISTS require_step_up boolean NOT NULL DEFAULT false;
