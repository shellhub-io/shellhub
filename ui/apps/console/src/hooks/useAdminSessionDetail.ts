import { useQuery } from "@tanstack/react-query";
import { getSessionAdminOptions } from "../client";
import { useAuthStore } from "../stores/authStore";
import { isSdkError } from "../api/errors";

/**
 * One session by UID, for the admin detail view. Cached for a minute only: a live session's
 * state changes while it is being looked at.
 */
export function useAdminSessionDetail(uid: string) {
  const isAdmin = useAuthStore((s) => s.isAdmin);

  const result = useQuery({
    ...getSessionAdminOptions({ path: { uid } }),
    enabled: isAdmin && !!uid,
    staleTime: 60 * 1000,
    retry: (count, err) => isSdkError(err) && err.status === 401 ? false : count < 1,
    refetchOnWindowFocus: false,
  });

  return {
    session: result.data ?? null,
    isLoading: result.isLoading,
    error: result.error,
  };
}
