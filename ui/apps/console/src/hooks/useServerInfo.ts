import { useQuery } from "@tanstack/react-query";
import { getInfoOptions } from "@/client";

/**
 * What the server reports about itself: its version, the endpoints users reach it on, and which
 * sign-in methods are enabled. The methods are an empty list until the server answers, so a
 * caller checks isLoading and isError before reading an empty list as none enabled.
 */
export function useServerInfo() {
  const { data, isLoading, isError } = useQuery(getInfoOptions());

  const signInMethods = [
    data?.authentication?.local && "Local",
    data?.authentication?.saml && "SAML",
  ].filter((method): method is string => !!method);

  return {
    version: data?.version,
    sshEndpoint: data?.endpoints?.ssh,
    apiEndpoint: data?.endpoints?.api,
    signInMethods,
    isLoading,
    isError,
  };
}
