import type { UserAuth } from "@/client";

/**
 * The localStorage key the auth store persists the session under.
 */
export const authSessionKey = "shellhub-session";

/**
 * The version of the persisted session. A seed written under another version is dropped on load.
 */
export const authSessionVersion = 1;

/**
 * The session a successful sign-in leaves in the auth store, read from the API's answer. Every
 * sign-in sets it, and the e2e suite seeds the persisted session with it, so a seeded browser
 * holds what signing in through the form would have left there.
 */
export function signedInSession(auth: UserAuth) {
  return {
    token: auth.token,
    user: auth.user,
    userId: auth.id,
    email: auth.email,
    tenant: auth.tenant,
    name: auth.name,
    isAdmin: auth.admin ?? false,
    mfaEnabled: auth.mfa || false,
    onboardingSurvey: auth.onboarding_survey,
    impersonated: auth.impersonated,
  };
}
