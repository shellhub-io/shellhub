import { vi } from "vitest";
import type { ClientConfig, Edition } from "../env";

export type { ClientConfig, Edition };

/**
 * The configuration a test starts from: community edition, every optional feature off, so a test
 * that needs one has to turn it on and says so.
 */
export const defaultConfig: ClientConfig = {
  version: "",
  edition: "community",
  announcements: false,
  webEndpoints: false,
  onboardingUrl: "",
  stripePublishableKey: "",
  chatwootWebsiteToken: "",
  chatwootBaseUrl: "",
};

/**
 * Stands in for the real loader. Resolves immediately with whatever getConfig returns, so a test
 * never waits on a fetch that will not happen.
 */
export const loadConfig = vi.fn(() => Promise.resolve(getConfig()));

/**
 * The configuration seen by the code under test. Override it to run a case against another
 * edition or feature flag.
 */
export const getConfig = vi.fn<() => ClientConfig>(() => defaultConfig);

/**
 * Whether the mocked configuration is community edition.
 */
export function isCommunity(): boolean {
  return getConfig().edition === "community";
}

/**
 * Whether the mocked configuration is enterprise edition.
 */
export function isEnterprise(): boolean {
  return getConfig().edition === "enterprise";
}

/**
 * Whether the mocked configuration is cloud edition.
 */
export function isCloud(): boolean {
  return getConfig().edition === "cloud";
}

/**
 * Whether the mocked edition has the paid features.
 */
export function isEnterpriseOrCloud(): boolean {
  return getConfig().edition !== "community";
}
