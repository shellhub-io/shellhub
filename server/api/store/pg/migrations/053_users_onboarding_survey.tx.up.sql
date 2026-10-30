ALTER TABLE users ADD COLUMN onboarding_survey character varying NOT NULL DEFAULT 'pending'
    CHECK (onboarding_survey IN ('pending', 'answered', 'skipped'));
