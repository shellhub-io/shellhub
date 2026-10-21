import { useMutation, useQuery } from "@tanstack/react-query";
import { adminDeleteUserMutation, getStatsOptions } from "@/client";
import { useAuthStore } from "@/stores/authStore";
import { isSdkError } from "@/api/errors";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";

/**
 * What every admin query shares: five minutes before it counts as stale, one retry except after a
 * 401, and no refetch when the window regains focus.
 */
export const adminQueryDefaults = {
  staleTime: 5 * 60 * 1000,
  retry: (count: number, err: unknown) =>
    isSdkError(err) && err.status === 401 ? false : count < 1,
  refetchOnWindowFocus: false,
};

/**
 * Instance-wide counts for the admin dashboard and the device limit banner. Idle unless the
 * signed-in user is an instance admin: the banner renders in the app layout for every enterprise
 * user, and only an admin may read the stats.
 */
export function useAdminStats() {
  const isAdmin = useAuthStore((s) => s.isAdmin);

  const result = useQuery({
    ...getStatsOptions(),
    ...adminQueryDefaults,
    enabled: isAdmin,
  });

  return {
    stats: result.data,
    isLoading: result.isLoading,
    isError: result.isError,
    error: result.error,
  };
}

/**
 * Deletes a user as an admin, along with the namespaces they own, refreshing the user and
 * namespace lists, the user's own query, the instance counts and
 * the namespace details that listed them as a member.
 */
export function useDeleteUser() {
  const invalidate = useInvalidateByIds(
    "getUsers",
    "getUser",
    "getStats",
    "getNamespaceAdmin",
    "getNamespacesAdmin",
  );
  return useMutation({
    ...adminDeleteUserMutation(),
    onSuccess: invalidate,
  });
}
