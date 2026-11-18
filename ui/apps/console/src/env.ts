import { nullOnFailure } from "@/utils/failure";

/**
 * Which edition the server is running. What each edition offers is decided by features() in
 * @/edition.
 */
export type Edition = "community" | "enterprise" | "cloud";

/**
 * The runtime configuration served as /config.json. It is read at startup rather than compiled
 * in, so one build of the console runs against any instance.
 */
export interface ClientConfig {
  version: string;
  edition: Edition;
  webEndpoints: boolean;
  onboardingUrl: string;
  stripePublishableKey: string;
  chatwootWebsiteToken: string;
  chatwootBaseUrl: string;
  mapStyle: string;
}

/**
 * Fallback used by `loadConfig` before `/config.json` resolves, and exported
 * for tests to spread as a known-good baseline. Production code should call
 * `getConfig()` so runtime overrides apply — never read this directly.
 */
export const defaultConfig: ClientConfig = {
  version: "",
  edition: "community",
  webEndpoints: false,
  onboardingUrl: "",
  stripePublishableKey: "",
  chatwootWebsiteToken: "",
  chatwootBaseUrl: "",
  mapStyle: "outline",
};

let cached: ClientConfig = defaultConfig;
let inflight: Promise<ClientConfig> | null = null;

/**
 * Fetches the runtime configuration once and caches it. Concurrent callers share one request,
 * and a failure leaves the defaults in place so a later call can retry — the console starts
 * either way rather than blocking on a config fetch.
 */
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

/**
 * The configuration as it stands, without waiting. Returns the defaults until loadConfig has
 * resolved, so anything that must not run against the defaults has to await loadConfig first.
 */
export function getConfig(): ClientConfig {
  return cached ?? defaultConfig;
}
