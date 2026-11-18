import { getConfig, type Edition } from "@/env";

/**
 * What the running edition offers. Each field is one product rule, decided here once, so a
 * feature asks for the rule it depends on instead of naming editions.
 *
 * - paid: multi-factor authentication, session recording, the server-side vault, firewall rules,
 *   web endpoints, the admin console, and namespaces created from the console. Without it,
 *   namespaces are created with the admin CLI.
 * - selfService: public sign-up, password reset and confirmation by email, invitations sent by
 *   email, and deleting one's own account.
 * - billing: the namespace subscription, meaning billing settings, the device chooser, the support
 *   chat, and a 402 that means the subscription refused the request.
 * - instanceLicense: the instance licence and its sign-in settings, meaning the licence pages and
 *   banner, SAML sign-in, and users awaiting approval.
 * - setupWizard: the first-run wizard that creates the instance's administrator.
 * - onboardingSurvey: the survey a new account answers on first sign-in.
 */
export interface EditionFeatures {
  paid: boolean;
  selfService: boolean;
  billing: boolean;
  instanceLicense: boolean;
  setupWizard: boolean;
  onboardingSurvey: boolean;
}

const FEATURES: Record<Edition, EditionFeatures> = {
  community: {
    paid: false,
    selfService: false,
    billing: false,
    instanceLicense: false,
    setupWizard: true,
    onboardingSurvey: true,
  },
  enterprise: {
    paid: true,
    selfService: false,
    billing: false,
    instanceLicense: true,
    setupWizard: true,
    onboardingSurvey: false,
  },
  cloud: {
    paid: true,
    selfService: true,
    billing: true,
    instanceLicense: false,
    setupWizard: false,
    onboardingSurvey: true,
  },
};

/**
 * The features of the edition getConfig reports. Like getConfig, it answers for the community
 * edition until loadConfig has resolved.
 */
export function features(): EditionFeatures {
  return FEATURES[getConfig().edition];
}
