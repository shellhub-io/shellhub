import { useQuery } from "@tanstack/react-query";
import { listConnections, getConnectionStatus } from "@/api/connections";

export function useConnections() {
  const result = useQuery({
    queryKey: ["connections"],
    queryFn: listConnections,
  });

  return {
    connections: result.data ?? [],
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}

/**
 * Reachability of a single connection. `online` is null while the first probe
 * is in flight. Re-probes periodically so the list stays roughly current.
 */
export function useConnectionStatus(id: string) {
  const result = useQuery({
    queryKey: ["connection-status", id],
    queryFn: () => getConnectionStatus(id),
    staleTime: 30_000,
    refetchInterval: 60_000,
  });

  return {
    online: result.isLoading ? null : (result.data ?? false),
    isChecking: result.isLoading,
  };
}
