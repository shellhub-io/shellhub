import { useQuery } from "@tanstack/react-query";
import {
  listConnections as listConnectionsSdk,
  type Connection as GeneratedConnection,
} from "@/client";
import { listConnectionsQueryKey } from "@/client/@tanstack/react-query.gen";
import { paginatedQueryFn, type PaginatedResult } from "@/api/pagination";
import { getConnectionStatus } from "@/api/connections";

export function useConnections({
  page = 1,
  perPage = 100,
}: { page?: number; perPage?: number } = {}) {
  const options = { query: { page, per_page: perPage } };

  const result = useQuery<PaginatedResult<GeneratedConnection>>({
    queryKey: listConnectionsQueryKey(options),
    queryFn: paginatedQueryFn(listConnectionsSdk, options),
  });

  return {
    connections: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
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
