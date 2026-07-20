/**
 * Dev-only onboarding-wizard simulation.
 *
 * Append `?wizard=demo` to any console URL in a dev build to force the
 * WelcomeWizard open regardless of the namespace's device state, then step
 * through it with a fake device, so the whole flow, including the final
 * "device connected" screen, can be validated without registering a real agent.
 *
 * Gated on `import.meta.env.DEV`, so the query param is inert in a production
 * build even if someone appends it.
 */
export const DEMO_DEVICE = {
  uid: "demo-device-0001",
  name: "raspberrypi-demo",
} as const;

export function isWizardDemo(): boolean {
  if (!import.meta.env.DEV) return false;
  return new URLSearchParams(window.location.search).get("wizard") === "demo";
}
