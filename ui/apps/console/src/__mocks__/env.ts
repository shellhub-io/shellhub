import { vi } from "vitest";
import type { ClientConfig, Edition } from "../env";

export type { ClientConfig, Edition };

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

export const loadConfig = vi.fn(() => Promise.resolve(getConfig()));

export const getConfig = vi.fn<() => ClientConfig>(() => defaultConfig);

export function isCommunity(): boolean {
  return getConfig().edition === "community";
}

export function isEnterprise(): boolean {
  return getConfig().edition === "enterprise";
}

export function isCloud(): boolean {
  return getConfig().edition === "cloud";
}

export function isEnterpriseOrCloud(): boolean {
  return getConfig().edition !== "community";
}
