import { useEffect } from "react";
import { useSearchParams } from "react-router-dom";
import {
  useQuery,
  type QueryKey,
  type UseQueryOptions,
} from "@tanstack/react-query";
import type { PaginatedResult } from "@/api/pagination";
import { pageCount } from "@/utils/pagination";
import { useDebouncedValue } from "./useDebouncedValue";
import {
  parseListParams,
  serializeListParams,
  type ListParamDefaults,
  type ListParamConstraints,
} from "./paginatedListParams";

/**
 * How a list's URL state is shaped: the defaults, which also define the keys, the constraints
 * applied to a value read from the URL, and a prefix for a page carrying more than one list.
 */
export interface ListStateConfig<T extends Record<string, unknown>> {
  defaults: ListParamDefaults<T>;
  constraints?: ListParamConstraints<T>;
  prefix?: string;
}

function prefixKey(key: string, prefix: string | undefined): string {
  return prefix ? `${prefix}.${key}` : key;
}

function stripPrefix<T extends Record<string, unknown>>(
  full: URLSearchParams,
  defaults: ListParamDefaults<T>,
  prefix: string | undefined,
): URLSearchParams {
  if (!prefix) return full;

  const stripped = new URLSearchParams();
  for (const key of Object.keys(defaults)) {
    for (const v of full.getAll(prefixKey(key, prefix))) {
      stripped.append(key, v);
    }
  }
  return stripped;
}

/**
 * The URL state of a paginated list without its query, for the lists not yet on usePagedList.
 */
export function useListState<T extends Record<string, unknown>>({
  defaults,
  constraints = {},
  prefix,
}: ListStateConfig<T>) {
  const [searchParams, setSearchParams] = useSearchParams();

  const params = parseListParams<T>(
    stripPrefix(searchParams, defaults, prefix),
    defaults,
    constraints,
  );

  const update = (patch: (current: T) => Record<string, unknown>) => {
    setSearchParams(
      (prev) => {
        const current = parseListParams<T>(
          stripPrefix(prev, defaults, prefix),
          defaults,
          constraints,
        );
        const merged = { ...current, ...patch(current) };
        const serialized = serializeListParams<T>(merged, defaults);

        const next = new URLSearchParams(prev);
        for (const key of Object.keys(defaults)) {
          const urlKey = prefixKey(key, prefix);
          next.delete(urlKey);
          for (const v of serialized.getAll(key)) next.append(urlKey, v);
        }
        return next;
      },
      { replace: true },
    );
  };

  return {
    params,
    setPage: (page: number) => update(() => ({ page })),
    setSearch: (search: string) => update(() => ({ search, page: defaults.page })),
    setFilter: <K extends keyof T>(key: K, value: T[K]) =>
      update(() => ({ [key]: value, page: defaults.page })),
    mapArrayFilter: <K extends keyof T>(key: K, fn: (current: T[K]) => T[K]) =>
      update((current) => ({ [key]: fn(current[key]), page: defaults.page })),
  };
}

const SEARCH_DEBOUNCE_MS = 300;

/**
 * A paginated list whose page, search and filters live in the URL, so a view is shareable
 * and survives a reload. Default-valued params stay out of the URL, params it does not manage are
 * kept, every setter but setPage returns to the first page, and a prefix namespaces the keys as
 * `<prefix>.<key>` so two lists can share a page.
 *
 * query builds the request from the params. When defaults has a search, query sees it trimmed and
 * 300 ms after the last keystroke, as does appliedSearch, while params.search is the raw input. Once a
 * response arrives with fewer pages than the current one, the list moves to its last page, or the
 * first when it is empty; it never moves while a request is pending or after one fails. Pages are
 * counted at PER_PAGE rows, so query must request that many.
 *
 * pagination carries the props a DataTable or Pagination pages with, to spread into it.
 */
export function usePagedList<
  T extends Record<string, unknown> & { page: number },
  TQueryFnData,
  TData extends PaginatedResult<unknown> = TQueryFnData & PaginatedResult<unknown>,
  TKey extends QueryKey = QueryKey,
>({
  query,
  ...config
}: ListStateConfig<T> & {
  query: (params: T) => UseQueryOptions<TQueryFnData, Error, TData, TKey>;
}) {
  const state = useListState(config);
  const { params, setPage } = state;

  const search = typeof params.search === "string" ? params.search.trim() : "";
  const debouncedSearch = useDebouncedValue(search, SEARCH_DEBOUNCE_MS);
  const queryParams =
    "search" in config.defaults ? { ...params, search: debouncedSearch } : params;

  const result = useQuery(query(queryParams));
  const totalCount = result.data?.totalCount ?? 0;
  const totalPages = pageCount(totalCount);
  const lastPage = Math.max(1, totalPages);
  const loaded = result.isSuccess && !result.isFetching;

  useEffect(() => {
    if (loaded && params.page > lastPage) setPage(lastPage);
  }, [loaded, params.page, lastPage, setPage]);

  return {
    ...state,
    appliedSearch: debouncedSearch,
    rows: (result.data?.data ?? []) as TData["data"],
    totalCount,
    totalPages,
    pagination: { page: params.page, totalPages, totalCount, onPageChange: setPage },
    isLoading: result.isLoading,
    error: result.error,
    refetch: result.refetch,
  };
}
