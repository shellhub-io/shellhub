/**
 * The titles of the setup steps, the same on the setup page and on the dashboard it hands over
 * to.
 */
export const SETUP_STEP_TITLES = {
  survey: "Tell us about you",
  account: "Create your account and namespace",
} as const;

/**
 * The titles of the device steps, the same on every screen that lists them.
 */
export const DEVICE_STEP_TITLES = {
  install: "Install the agent",
  pair: "Check it's your device, then pair",
  shell: "Open a shell on it",
} as const;
