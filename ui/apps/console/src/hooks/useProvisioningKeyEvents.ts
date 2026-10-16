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
 * What the pending queue holds for the current member: whether they may work it at all, whether
 * it is worth linking (always in a legacy namespace, in a new one once a device registered with
 * the tenant ID), how many of the newest hundred still wait for a decision, and whether it lists
 * every one of a count of pending devices. In a new namespace a device enrolled by a provisioning
 * key also waits pending and the queue does not list it, so `holds` is false whenever the count
 * exceeds what the queue has waiting. A member who may not accept devices gets none of it and
 * nothing is asked.
 */
export function useTenantRegistrationQueue() {
  const allowed = useHasPermission("device:accept");
  const legacy = useNamespaceLegacy() === true;
  const { events, totalCount } = useTenantRegistrations({ enabled: allowed });

  const reachable = allowed && (legacy || totalCount > 0);
  const waiting = events.filter(
    (event) => event.is_current && event.device_status === "pending",
  ).length;

  return {
    allowed,
    reachable,
    waiting,
    holds: (pendingDevices: number) =>
      reachable && (legacy || pendingDevices <= waiting),
  };
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
