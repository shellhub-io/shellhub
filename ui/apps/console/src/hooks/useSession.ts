import { useQuery } from "@tanstack/react-query";
import { getSessionOptions } from "../client";

/**
 * One session by UID. Idle until a UID is given.
 */
export function useSession(uid: string) {
  const result = useQuery({
    ...getSessionOptions({ path: { uid } }),
    enabled: !!uid,
  });

  return {
    session: result.data ?? null,
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}
