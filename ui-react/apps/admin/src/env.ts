export interface ClientConfig {
  version: string;
  enterprise: boolean;
  cloud: boolean;
}

const defaultConfig: ClientConfig = {
  version: "",
  enterprise: false,
  cloud: false,
};

let cached: ClientConfig = defaultConfig;

export async function loadConfig(): Promise<ClientConfig> {
  if (cached !== defaultConfig) return cached;

  try {
    const res = await fetch("/v2/ui/config.json");
    cached = { ...defaultConfig, ...(await res.json()) };
  } catch {
    cached = defaultConfig;
  }

  return cached;
}

export function getConfig(): ClientConfig {
  return cached ?? defaultConfig;
}
