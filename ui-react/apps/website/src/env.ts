/// <reference types="vite/client" />

declare global {
  interface Window {
    env?: Record<string, string>;
  }
}

function getEnv(key: string): string {
  return window.env?.[key] ?? import.meta.env[key] ?? "";
}

export const env = {
  version: getEnv("VITE_SHELLHUB_SHELLHUB_VERSION"),
  isEnterprise: getEnv("VITE_SHELLHUB_SHELLHUB_ENTERPRISE") === "true",
  isCloud: getEnv("VITE_SHELLHUB_SHELLHUB_CLOUD") === "true",
  sentryDsn: getEnv("VITE_SHELLHUB_SHELLHUB_SENTRY_DSN"),
};
