/* eslint-disable */
// import.meta.env.MODE
export const envVariables = {
  isEnterprise: process.env.VUE_APP_SHELLHUB_ENTERPRISE === "true",
  isCloud: process.env.VUE_APP_SHELLHUB_CLOUD === "true",
  stripePublishableKey: process.env.VUE_APP_SHELLHUB_STRIPE_PUBLISHABLE_KEY,
  billingEnable: process.env.VUE_APP_SHELLHUB_BILLING === "true",
  version: process.env.SHELLHUB_VERSION,
  announcementsEnable: process.env.VUE_APP_SHELLHUB_ANNOUNCEMENTS === "true",
  stripeKey: process.env.VUE_APP_SHELLHUB_STRIPE_PUBLISHABLE_KEY,
};
