import { useInfiniteQuery } from "@tanstack/react-query";
import {
  provisioningKeyHistory,
  provisioningKeyHistoryQueryKey,
  type ProvisioningKeyHistoryData,
  type ProvisioningKeyEvent,
} from "../client";
import { paginatedQueryFn } from "../api/pagination";

interface UseProvisioningKeyEventsParams {
  id: string | null;
  perPage?: number;
}

/**
 * The enrolments made with a provisioning key, newest first, a page at a time: loadMore appends
 * the next page to events until hasMore turns false. Pages are offsets, so a registration landing
 * between two fetches shifts the rows and repeats one across the boundary; events keeps each id
 * once. loadMoreFailed tells a failed later page apart from a failed first load.
 */
export function useProvisioningKeyEvents({
  id,
  perPage = 100,
}: UseProvisioningKeyEventsParams) {
  const base = {
    path: { id: id ?? "" },
    query: { per_page: perPage, sort_by: "created_at", order_by: "desc" },
  } satisfies {
    path: ProvisioningKeyHistoryData["path"];
    query: ProvisioningKeyHistoryData["query"];
  };

  const result = useInfiniteQuery({
    queryKey: provisioningKeyHistoryQueryKey(base),
    queryFn: ({ pageParam }) =>
      paginatedQueryFn(provisioningKeyHistory, {
        ...base,
        query: { ...base.query, page: pageParam },
      })(),
    initialPageParam: 1,
    getNextPageParam: (last, pages) =>
      pages.length * perPage < last.totalCount ? pages.length + 1 : undefined,
    enabled: !!id,
  });

  const pages = result.data?.pages ?? [];
  const events: ProvisioningKeyEvent[] = [
    ...new Map(pages.flatMap((p) => p.data).map((e) => [e.id, e])).values(),
  ];

  return {
    events,
    totalCount: pages[0]?.totalCount ?? 0,
    hasMore: result.hasNextPage,
    loadMore: () => void result.fetchNextPage(),
    isLoadingMore: result.isFetchingNextPage,
    loadMoreFailed: result.isFetchNextPageError,
    isLoading: result.isLoading,
    error: result.error,
  };
}
