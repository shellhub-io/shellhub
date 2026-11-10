import { useSearchParams } from "react-router-dom";
import {
  keepPreviousData,
  useInfiniteQuery,
  type QueryKey,
} from "@tanstack/react-query";
import {
  listTenantRegistrations,
  listTenantRegistrationsQueryKey,
  provisioningKeyHistory,
  provisioningKeyHistoryQueryKey,
  type ProvisioningKeyHistoryData,
  type ProvisioningKeyEvent,
} from "../client";
import { paginatedQueryFn, type PaginatedResult } from "../api/pagination";
import { useDebouncedValue } from "./useDebouncedValue";
import { useHasPermission } from "./useHasPermission";
import { useNamespaceLegacy } from "./useNamespaceLegacy";

/**
 * The orders a registration history can be read in, by the key ?sort= carries.
 */
export const REGISTRATION_SORTS = {
  newest: { label: "Newest first", sort_by: "created_at", order_by: "desc" },
  oldest: { label: "Oldest first", sort_by: "created_at", order_by: "asc" },
} as const;

/** An order a registration history can be read in; "newest" is the default and is left out of ?sort=. */
export type RegistrationSort = keyof typeof REGISTRATION_SORTS;

/**
 * What narrows and orders a registration history: the hostname text to search for, blank for none,
 * and the order. useRegistrationView returns one.
 */
export interface RegistrationQuery {
  hostname: string;
  sort: RegistrationSort;
}

const ALL_NEWEST_FIRST: RegistrationQuery = { hostname: "", sort: "newest" };

interface UseProvisioningKeyEventsParams {
  id: string | null;
  perPage?: number;
  view?: RegistrationQuery;
}

const SEARCH_PARAM = "search";
const SORT_PARAM = "sort";
const SEARCH_DEBOUNCE_MS = 300;

function isRegistrationSort(value: string | null): value is RegistrationSort {
  return value !== null && Object.hasOwn(REGISTRATION_SORTS, value);
}

function queryOf({ hostname, sort }: RegistrationQuery) {
  const { sort_by, order_by } = REGISTRATION_SORTS[sort];
  return {
    sort_by,
    order_by,
    ...(hostname ? { "filters[hostname][contains]": hostname } : {}),
  };
}

/**
 * How a registration history is read: its search by hostname and its order, kept in the URL as
 * ?search= and ?sort= so a view can be shared. search is the raw input for the field; hostname is
 * it trimmed and 300 ms after the last keystroke, what the history is narrowed by. sort is newest
 * when ?sort= is absent or names no order of REGISTRATION_SORTS. A blank search and the newest
 * order are left out of the URL.
 */
export function useRegistrationView() {
  const [params, setParams] = useSearchParams();
  const search = params.get(SEARCH_PARAM) ?? "";
  const hostname = useDebouncedValue(search.trim(), SEARCH_DEBOUNCE_MS);
  const sortParam = params.get(SORT_PARAM);
  const sort: RegistrationSort = isRegistrationSort(sortParam)
    ? sortParam
    : "newest";

  const setParam = (name: string, value: string | null) => {
    setParams(
      (prev) => {
        const updated = new URLSearchParams(prev);
        if (value) updated.set(name, value);
        else updated.delete(name);
        return updated;
      },
      { replace: true },
    );
  };

  return {
    search,
    setSearch: (next: string) =>
      setParam(SEARCH_PARAM, next.trim() ? next : null),
    hostname,
    sort,
    setSort: (next: RegistrationSort) =>
      setParam(SORT_PARAM, next === "newest" ? null : next),
  };
}

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
    placeholderData: keepPreviousData,
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
    isRefreshing: result.isPlaceholderData,
    error: result.error,
  };
}

/**
 * A history of registrations, in the order asked for, a page at a time: loadMore appends the next page to
 * events until hasMore turns false. Pages are offsets, so a registration landing between two
 * fetches shifts the rows and repeats one across the boundary; events keeps each id once.
 * loadMoreFailed tells a failed later page apart from a failed first load. While a new search or
 * order loads, events keeps the previous one's and isRefreshing is true.
 */
export type RegistrationHistory = ReturnType<typeof useRegistrationPages>;

/**
 * The enrolments made with a provisioning key, as a RegistrationHistory narrowed and ordered by
 * view, all of them newest first without one.
 */
export function useProvisioningKeyEvents({
  id,
  perPage = 100,
  view = ALL_NEWEST_FIRST,
}: UseProvisioningKeyEventsParams): RegistrationHistory {
  const base = {
    path: { id: id ?? "" },
    query: { per_page: perPage, ...queryOf(view) },
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
 * The registrations of the namespace's tenant-only key, as a RegistrationHistory narrowed and
 * ordered by view, all of them newest first without one. It reads them from the device API, so any
 * member who may accept devices gets them.
 */
export function useTenantRegistrations({
  perPage = 100,
  enabled = true,
  view = ALL_NEWEST_FIRST,
}: {
  perPage?: number;
  enabled?: boolean;
  view?: RegistrationQuery;
} = {}): RegistrationHistory {
  const base = {
    query: { per_page: perPage, ...queryOf(view) },
  };

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
