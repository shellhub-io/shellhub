-- Users provisioned on first login through an identity provider brokered by the embedded OIDC
-- provider. This is distinct from 'saml', which names the older single-IdP service provider that
-- ShellHub implements itself; both can be in use on the same instance during the transition.
--
-- This file is deliberately not transactional (no .tx. in its name): a value added to an enum
-- inside a transaction cannot be referenced until that transaction commits, and running these
-- outside one keeps the migration valid regardless of what a later statement does.
ALTER TYPE user_origin ADD VALUE IF NOT EXISTS 'external';

--bun:split

ALTER TYPE user_auth_method ADD VALUE IF NOT EXISTS 'external';
