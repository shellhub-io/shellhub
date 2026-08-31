import { useQuery } from "@tanstack/react-query";
import { getContainerOptions } from "../client";

/**
 * One container by UID. Idle until a UID is given, so a route that has not resolved its
 * parameter yet does not issue a request for an empty path.
 */
export function useContainer(uid: string) {
  const result = useQuery({
    ...getContainerOptions({ path: { uid } }),
    enabled: !!uid,
  });

  return {
    container: (result.data ?? null),
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}
