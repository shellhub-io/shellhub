import { useQuery } from "@tanstack/react-query";
import { getInfoOptions } from "@/client";

import { DEFAULT_SSH_PORT } from "@/utils/sshid";

/**
 * Where users reach devices over SSH, as the server reports it: its host and port. Falls back to
 * this page's host and the default port until the server answers, or when it reports no SSH
 * endpoint.
 */
export function useSshEndpoint(): { host: string; port: number } {
  const { data } = useQuery(getInfoOptions());
  const endpoint = data?.endpoints?.ssh ?? "";
  const colon = endpoint.lastIndexOf(":");
  const host =
    (colon < 0 ? endpoint : endpoint.slice(0, colon)) ||
    window.location.hostname;
  const port = colon < 0 ? NaN : Number(endpoint.slice(colon + 1));

  return {
    host,
    port: Number.isInteger(port) && port > 0 ? port : DEFAULT_SSH_PORT,
  };
}
