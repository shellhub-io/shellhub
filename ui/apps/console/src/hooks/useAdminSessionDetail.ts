import { useQuery } from "@tanstack/react-query";
import { getSessionAdminOptions } from "../client/@tanstack/react-query.gen";
import { useAuthStore } from "../stores/authStore";
import { isSdkError } from "../api/errors";

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
