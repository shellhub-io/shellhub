import { useQuery } from "@tanstack/react-query";
import { getUserInfoOptions } from "@/client";

/**
 * The signed-in user's account as the server holds it, namespace creation limit included.
 * Idle while `enabled` is false.
 */
export function useUserInfo({ enabled = true }: { enabled?: boolean } = {}) {
  const result = useQuery({ ...getUserInfoOptions(), enabled });

  return {
    user: result.data ?? null,
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}
