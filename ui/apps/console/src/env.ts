import { nullOnFailure } from "@/utils/failure";

export type Edition = "community" | "enterprise" | "cloud";

export interface ClientConfig {
  version: string;
  edition: Edition;
  announcements: boolean;
  webEndpoints: boolean;
  onboardingUrl: string;
  stripePublishableKey: string;
  chatwootWebsiteToken: string;
  chatwootBaseUrl: string;
}

/**
 * Fallback used by `loadConfig` before `/config.json` resolves, and exported
 * for tests to spread as a known-good baseline. Production code should call
 * `getConfig()` so runtime overrides apply — never read this directly.
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

let cached: ClientConfig = defaultConfig;
let inflight: Promise<ClientConfig> | null = null;

export async function loadConfig(): Promise<ClientConfig> {
  if (cached !== defaultConfig) return cached;
  if (inflight) return inflight;

  inflight = (async () => {
    const fetched = await fetch("/config.json")
      .then((res) => res.json() as Promise<Partial<ClientConfig>>)
      .catch(nullOnFailure);

    cached = fetched ? { ...defaultConfig, ...fetched } : defaultConfig;
    inflight = null;

    return cached;
  })();

  return inflight;
}

export function getConfig(): ClientConfig {
  return cached ?? defaultConfig;
}

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
