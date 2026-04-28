export interface ClientConfig {
  version: string;
  enterprise: boolean;
  cloud: boolean;
  announcements: boolean;
  webEndpoints: boolean;
  onboardingUrl: string;
}

const defaultConfig: ClientConfig = {
  version: "",
  enterprise: false,
  cloud: false,
  announcements: false,
  webEndpoints: false,
  onboardingUrl: "",
};

let cached: ClientConfig = defaultConfig;

export async function loadConfig(): Promise<ClientConfig> {
  if (cached !== defaultConfig) return cached;

  try {
    const res = await fetch("/config.json");
    cached = { ...defaultConfig, ...(await res.json() as Partial<ClientConfig>) };
  } catch {
    cached = defaultConfig;
  }

  return cached;
}

export function getConfig(): ClientConfig {
  return cached ?? defaultConfig;
}
