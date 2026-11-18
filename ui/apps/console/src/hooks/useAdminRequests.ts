import { useMutation, useQuery } from "@tanstack/react-query";
import { adminDeleteUserMutation, getStatsOptions } from "@/client";
import { useAuthStore } from "@/stores/authStore";
import { adminQueryDefaults } from "@/hooks/adminQueryDefaults";
import { useInvalidateByIds } from "@/hooks/useInvalidateQueries";

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
