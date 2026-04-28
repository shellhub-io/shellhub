export interface ClientConfig {
  version: string;
  enterprise: boolean;
  cloud: boolean;
  announcements: boolean;
  webEndpoints: boolean;
  onboardingUrl: string;
  stripePublishableKey: string;
  chatwootWebsiteToken: string;
  chatwootBaseUrl: string;
}

const defaultConfig: ClientConfig = {
  version: "",
  enterprise: false,
  cloud: false,
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
    try {
      const res = await fetch("/config.json");
      cached = {
        ...defaultConfig,
        ...((await res.json()) as Partial<ClientConfig>),
      };
    } catch {
      // leave cached as defaultConfig so future calls can retry
    } finally {
      inflight = null;
    }
    return cached;
  })();

  return inflight;
}

export function getConfig(): ClientConfig {
  return cached ?? defaultConfig;
}
