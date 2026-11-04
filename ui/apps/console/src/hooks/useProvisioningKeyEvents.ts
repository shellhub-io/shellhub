import { useInfiniteQuery, type QueryKey } from "@tanstack/react-query";
import {
  listTenantRegistrations,
  listTenantRegistrationsQueryKey,
  provisioningKeyHistory,
  provisioningKeyHistoryQueryKey,
  type ProvisioningKeyHistoryData,
  type ProvisioningKeyEvent,
} from "../client";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";
import { useHasPermission } from "./useHasPermission";
import { useNamespaceLegacy } from "./useNamespaceLegacy";

interface UseProvisioningKeyEventsParams {
  id: string | null;
  perPage?: number;
}

const newestFirst = { sort_by: "created_at", order_by: "desc" } as const;

function useRegistrationPages(
  queryKey: QueryKey,
  fetchPage: (page: number) => Promise<PaginatedResult<ProvisioningKeyEvent>>,
  perPage: number,
  enabled: boolean,
) {
  const result = useInfiniteQuery({
    queryKey,
    queryFn: ({ pageParam }) => fetchPage(pageParam),
    initialPageParam: 1,
    getNextPageParam: (last, pages) =>
      pages.length * perPage < last.totalCount ? pages.length + 1 : undefined,
    enabled,
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

/**
 * A history of registrations, newest first, a page at a time: loadMore appends the next page to
 * events until hasMore turns false. Pages are offsets, so a registration landing between two
 * fetches shifts the rows and repeats one across the boundary; events keeps each id once.
 * loadMoreFailed tells a failed later page apart from a failed first load.
 */
export type RegistrationHistory = ReturnType<typeof useRegistrationPages>;

/**
 * The enrolments made with a provisioning key, as a RegistrationHistory.
 */
export function useProvisioningKeyEvents({
  id,
  perPage = 100,
}: UseProvisioningKeyEventsParams): RegistrationHistory {
  const base = {
    path: { id: id ?? "" },
    query: { per_page: perPage, ...newestFirst },
  } satisfies {
    path: ProvisioningKeyHistoryData["path"];
    query: ProvisioningKeyHistoryData["query"];
  };

  return useRegistrationPages(
    provisioningKeyHistoryQueryKey(base),
    (page) =>
      paginatedQueryFn(provisioningKeyHistory, {
        ...base,
        query: { ...base.query, page },
      })(),
    perPage,
    !!id,
  );
}

/**
 * The tenant ID queue as the current member sees it: whether they may accept devices, whether the
 * queue is reachable, which is only in a legacy namespace and only for a member who may accept
 * devices, and how many of its newest hundred registrations still wait for a decision. A new
 * namespace sends members to the provisioning keys instead, so it asks for no registrations.
 */
export function useTenantRegistrationQueue() {
  const allowed = useHasPermission("device:accept");
  const legacy = useNamespaceLegacy() === true;
  const reachable = allowed && legacy;
  const { events } = useTenantRegistrations({ enabled: reachable });

  const waiting = events.filter(
    (event) => event.is_current && event.device_status === "pending",
  ).length;

  return { allowed, reachable, waiting };
}

/**
 * The registrations of the namespace's tenant-only key, as a RegistrationHistory. It reads them
 * from the device API, so any member who may accept devices gets them.
 */
export function useTenantRegistrations({
  perPage = 100,
  enabled = true,
}: { perPage?: number; enabled?: boolean } = {}): RegistrationHistory {
  const base = { query: { per_page: perPage, ...newestFirst } };

  return useRegistrationPages(
    listTenantRegistrationsQueryKey(base),
    (page) =>
      paginatedQueryFn(listTenantRegistrations, {
        query: { ...base.query, page },
      })(),
    perPage,
    enabled,
  );
}
