import { useCallback, useRef } from "react";
import { useSearchParams } from "react-router-dom";
import {
  parseListParams,
  serializeListParams,
  type ListParamDefaults,
  type ListParamConstraints,
} from "./paginatedListParams";

/**
 * Describes a sortable field: which key it maps to and what order to use when
 * it is first selected (i.e. when the user switches from a different field).
 */
export interface SortFieldDef {
  field: string;
  initialOrder: "asc" | "desc";
}

export interface UsePaginatedListStateConfig<T extends Record<string, unknown>> {
  defaults: ListParamDefaults<T>;
  constraints?: ListParamConstraints<T>;
  sortFields?: SortFieldDef[];
  prefix?: string;
}

export interface UsePaginatedListStateResult<T extends Record<string, unknown>> {
  params: T;
  searchString: string;
  setPage: (page: number) => void;
  setSearch: (search: string) => void;
  setFilter: <K extends keyof T>(key: K, value: T[K]) => void;
  setArrayFilter: <K extends keyof T>(key: K, values: T[K]) => void;
  mapArrayFilter: <K extends keyof T>(
    key: K,
    fn: (current: T[K]) => T[K],
  ) => void;
  handleSort: (field: string) => void;
  reset: () => void;
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
 * Stable empty-constraints sentinel.  Using a module-level constant instead of
 * an inline `{}` default means callers that don't supply `constraints` all
 * share the same object reference, keeping the `update` useCallback stable
 * across renders.
 */
const EMPTY_CONSTRAINTS: ListParamConstraints<Record<string, unknown>> = {};

/**
 * Config-driven hook that syncs a paginated list's URL state (page, search,
 * sort, and any extra dimensions) with `useSearchParams`.
 *
 * Rules:
 * - Default-valued params are omitted from the URL (replace-history mode).
 * - Unrelated params already present in the URL are always preserved.
 * - Every non-page setter resets page to 1.
 * - When a `prefix` is given, every managed key is namespaced as
 *   `<prefix>.<key>` in the URL so multiple instances coexist cleanly.
 */
export function usePaginatedListState<T extends Record<string, unknown>>({
  defaults,
  constraints = EMPTY_CONSTRAINTS as ListParamConstraints<T>,
  sortFields,
  prefix,
}: UsePaginatedListStateConfig<T>): UsePaginatedListStateResult<T> {
  const [searchParams, setSearchParams] = useSearchParams();

  const paramsRef = useRef<T | undefined>(undefined);
  const stripped = stripPrefix(searchParams, defaults, prefix);
  // eslint-disable-next-line react-hooks/refs -- caches prior parse result for referential stability
  const params = parseListParams<T>(stripped, defaults, constraints, paramsRef.current);
  // eslint-disable-next-line react-hooks/refs -- stash this render's result for the next render
  paramsRef.current = params;

  const update = useCallback(
    (patch: Partial<T>) => {
      setSearchParams(
        (prev) => {
          const next = new URLSearchParams(prev);

          const currentStripped = stripPrefix(prev, defaults, prefix);
          const current = parseListParams<T>(currentStripped, defaults, constraints);
          const merged = { ...current, ...patch };

          const serialized = serializeListParams<T>(merged, defaults);

          for (const key of Object.keys(defaults)) {
            const urlKey = prefixKey(key, prefix);
            next.delete(urlKey);
            const newValues = serialized.getAll(key);
            for (const v of newValues) {
              next.append(urlKey, v);
            }
          }

          return next;
        },
        { replace: true },
      );
    },
    [setSearchParams, defaults, constraints, prefix],
  );

  const setPage = useCallback(
    (page: number) => {
      update({ page } as unknown as Partial<T>);
    },
    [update],
  );

  const setSearch = useCallback(
    (search: string) => {
      update({ search, page: defaults.page } as unknown as Partial<T>);
    },
    [update, defaults.page],
  );

  const setFilter = useCallback(
    <K extends keyof T>(key: K, value: T[K]) => {
      update({ [key]: value, page: defaults.page } as unknown as Partial<T>);
    },
    [update, defaults.page],
  );

  const setArrayFilter = useCallback(
    <K extends keyof T>(key: K, values: T[K]) => {
      update({ [key]: values, page: defaults.page } as unknown as Partial<T>);
    },
    [update, defaults.page],
  );

  const mapArrayFilter = useCallback(
    <K extends keyof T>(key: K, fn: (current: T[K]) => T[K]) => {
      setSearchParams(
        (prev) => {
          const currentStripped = stripPrefix(prev, defaults, prefix);
          const current = parseListParams<T>(currentStripped, defaults, constraints);
          const next = fn(current[key]);
          const merged = { ...current, [key]: next, page: defaults.page };
          const serialized = serializeListParams<T>(merged, defaults);

          const nextParams = new URLSearchParams(prev);
          for (const k of Object.keys(defaults)) {
            const urlKey = prefixKey(k, prefix);
            nextParams.delete(urlKey);
            const newValues = serialized.getAll(k);
            for (const v of newValues) {
              nextParams.append(urlKey, v);
            }
          }
          return nextParams;
        },
        { replace: true },
      );
    },
    [setSearchParams, defaults, constraints, prefix],
  );

  const handleSort = useCallback(
    (field: string) => {
      setSearchParams(
        (prev) => {
          const currentStripped = stripPrefix(prev, defaults, prefix);
          const current = parseListParams<T>(currentStripped, defaults, constraints);
          const p = current as Record<string, unknown>;
          const currentField = p["sortField"] as string | undefined;
          const currentOrder = p["sortOrder"] as "asc" | "desc" | undefined;

          const nextOrder: "asc" | "desc" =
            field === currentField
              ? // Same field — toggle.
                currentOrder === "asc"
                ? "desc"
                : "asc"
              : // New field — use its declared initialOrder, falling back to "asc".
                (sortFields?.find((f) => f.field === field)?.initialOrder ?? "asc");

          const patch = {
            sortField: field,
            sortOrder: nextOrder,
            page: defaults.page,
          } as unknown as Partial<T>;

          const merged = { ...current, ...patch };
          const serialized = serializeListParams<T>(merged, defaults);

          const next = new URLSearchParams(prev);
          for (const key of Object.keys(defaults)) {
            const urlKey = prefixKey(key, prefix);
            next.delete(urlKey);
            const newValues = serialized.getAll(key);
            for (const v of newValues) {
              next.append(urlKey, v);
            }
          }
          return next;
        },
        { replace: true },
      );
    },
    [setSearchParams, defaults, constraints, prefix, sortFields],
  );

  const reset = useCallback(() => {
    update({ ...defaults });
  }, [update, defaults]);

  return {
    params,
    searchString: searchParams.toString(),
    setPage,
    setSearch,
    setFilter,
    setArrayFilter,
    mapArrayFilter,
    handleSort,
    reset,
  };
}
