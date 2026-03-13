import type { CreateClientConfig } from "../client/client.gen";

export const createClientConfig: CreateClientConfig = (config) => ({
  ...config,
  baseUrl: `${window.location.protocol}//${window.location.host}`,
});
