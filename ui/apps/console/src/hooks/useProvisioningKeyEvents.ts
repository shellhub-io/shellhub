import { useQuery } from "@tanstack/react-query";
import {
  provisioningKeyHistory,
  provisioningKeyHistoryQueryKey,
  type ProvisioningKeyHistoryData,
  type ProvisioningKeyEvent,
} from "../client";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";

interface UseProvisioningKeyEventsParams {
  id: string | null;
  page?: number;
  perPage?: number;
}

/**
 * A page of the enrolments made with a provisioning key — what it was used for, and when.
 */
export function useProvisioningKeyEvents({
  id,
  page = 1,
  perPage = 15,
}: UseProvisioningKeyEventsParams) {
  const options = {
    path: { id: id ?? "" },
    query: { page, per_page: perPage, sort_by: "created_at", order_by: "desc" },
  } satisfies {
    path: ProvisioningKeyHistoryData["path"];
    query: ProvisioningKeyHistoryData["query"];
  };

  const result = useQuery<PaginatedResult<ProvisioningKeyEvent>>({
    queryKey: provisioningKeyHistoryQueryKey(options),
    queryFn: paginatedQueryFn(provisioningKeyHistory, options),
    enabled: !!id,
  });

  return {
    events: result.data?.data ?? [],
    totalCount: result.data?.totalCount ?? 0,
    isLoading: result.isLoading,
    error: result.error,
  };
}
