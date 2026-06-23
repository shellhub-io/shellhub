import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import apiClient from "@/api/client";
import type { Share } from "@/types/share";

// Active shares live in-memory in the ssh service, so we poll for fresh viewer counts.
const REFETCH_INTERVAL_MS = 4000;

export function useShares() {
  const result = useQuery<Share[]>({
    queryKey: ["shares"],
    queryFn: async () => {
      const res = await apiClient.get<Share[]>("/ssh/shares");
      return res.data ?? [];
    },
    refetchInterval: REFETCH_INTERVAL_MS,
  });

  return {
    shares: result.data ?? [],
    isLoading: result.isLoading,
    error: result.error,
  };
}

export function useRevokeShare() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: async (token: string) => {
      await apiClient.delete(`/ssh/shares/${token}`);
    },
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: ["shares"] });
    },
  });
}
