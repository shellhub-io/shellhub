import type { CreateClientConfig } from "../client/client.gen";

/**
 * Points the generated SDK at the origin serving the console. The API is behind the same gateway
 * as the UI, so there is no separate host to configure and none to get wrong.
 */
export const createClientConfig: CreateClientConfig = (config) => ({
  ...config,
  baseUrl: `${window.location.protocol}//${window.location.host}`,
});
