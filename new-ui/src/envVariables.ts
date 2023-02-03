/* eslint-disable */
// @ts-nocheck

export const envVariables = {
  isEnterprise: (window.env || process.env).VUE_APP_SHELLHUB_ENTERPRISE === "true",
  isCloud: (window.env || process.env).VUE_APP_SHELLHUB_CLOUD === "true",
  stripePublishableKey: (window.env || process.env).VUE_APP_SHELLHUB_STRIPE_PUBLISHABLE_KEY,
  billingEnable: (window.env || process.env).VUE_APP_SHELLHUB_BILLING === "true",
  version: (window.env || process.env).SHELLHUB_VERSION,
  announcementsEnable: (window.env || process.env).VUE_APP_SHELLHUB_ANNOUNCEMENTS === "true",
  stripeKey: (window.env || process.env).VUE_APP_SHELLHUB_STRIPE_PUBLISHABLE_KEY,
};
