import { useQuery } from "@tanstack/react-query";
import {
  installKeyHistory,
  installKeyHistoryQueryKey,
  type InstallKeyHistoryData,
  type InstallKeyEvent,
} from "../client";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";

interface UseInstallKeyEventsParams {
  id: string | null;
  page?: number;
  perPage?: number;
}

export function useInstallKeyEvents({
  id,
  page = 1,
  perPage = 15,
}: UseInstallKeyEventsParams) {
  const options = {
    path: { id: id ?? "" },
    query: { page, per_page: perPage, sort_by: "created_at", order_by: "desc" },
  } satisfies {
    path: InstallKeyHistoryData["path"];
    query: InstallKeyHistoryData["query"];
  };

  const result = useQuery<PaginatedResult<InstallKeyEvent>>({
    queryKey: installKeyHistoryQueryKey(options),
    queryFn: paginatedQueryFn(installKeyHistory, options),
    enabled: !!id,
  });

  return {
    events: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
  };
}
